package s3

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/KARTIKrocks/objstore"
)

// fakeS3 is a minimal S3-compatible HTTP server covering just enough of the
// PutObject / multipart-upload surface to exercise manager.Uploader against
// a real (loopback) HTTP round trip, without depending on a live S3/MinIO
// instance. Path-style addressing only, matching how Storage.New configures
// the client for custom endpoints.
type fakeS3 struct {
	mu                  sync.Mutex
	uploadID            string
	failPartNumber      string // if set, UploadPart for this partNumber always fails
	failAbortAttempts   int    // this many leading AbortMultipartUpload calls fail
	completeIfNoneMatch string
	abortAttempts       int           // every AbortMultipartUpload call, success or not
	abortRequests       []string      // uploadIds seen on a *successful* AbortMultipartUpload
	partHeaders         []http.Header // headers seen on each UploadPart request, in order
}

func (f *fakeS3) server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f.handle))
}

func (f *fakeS3) handle(w http.ResponseWriter, r *http.Request) {
	// Drain the request body before responding. Writing a response while the
	// client is still sending a large body (a 5MiB+ part) races the server's
	// connection handling and can surface as a spurious "connection reset by
	// peer" on the client side, which the SDK then retries into a hard
	// failure instead of the deterministic outcome each test expects.
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}

	q := r.URL.Query()

	switch {
	case r.Method == http.MethodPost && q.Has("uploads"):
		f.mu.Lock()
		f.uploadID = "test-upload-id"
		id := f.uploadID
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<InitiateMultipartUploadResult><Bucket>test-bucket</Bucket><Key>obj.bin</Key><UploadId>%s</UploadId></InitiateMultipartUploadResult>`, id)

	case r.Method == http.MethodPut && q.Has("uploadId") && q.Has("partNumber"):
		partNumber := q.Get("partNumber")
		f.mu.Lock()
		f.partHeaders = append(f.partHeaders, r.Header.Clone())
		shouldFail := f.failPartNumber == partNumber
		f.mu.Unlock()
		if shouldFail {
			// A permanent (not one-shot) failure, since the SDK retries
			// transient errors — a single failed attempt wouldn't
			// deterministically fail the overall upload.
			writeS3Error(w, http.StatusBadRequest, "InvalidPart", "simulated permanent part failure")
			return
		}
		w.Header().Set("ETag", `"part-etag"`)
		w.WriteHeader(http.StatusOK)

	case r.Method == http.MethodPost && q.Has("uploadId"):
		f.mu.Lock()
		f.completeIfNoneMatch = r.Header.Get("If-None-Match")
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadResult><Location>http://example.com/obj.bin</Location><Bucket>test-bucket</Bucket><Key>obj.bin</Key><ETag>"final-etag-2"</ETag></CompleteMultipartUploadResult>`)

	case r.Method == http.MethodDelete && q.Has("uploadId"):
		f.mu.Lock()
		f.abortAttempts++
		failThisOne := f.abortAttempts <= f.failAbortAttempts
		f.mu.Unlock()
		if failThisOne {
			// Simulates manager.Uploader's own abort call failing — e.g.
			// because it reused the same context that caused the upload to
			// fail in the first place (a cancellation or deadline). Failing
			// enough leading attempts exhausts the SDK's own retry budget for
			// that logical call, so only a distinct, later call (Put's
			// abortOrphanedMultipart fallback) can still succeed.
			writeS3Error(w, http.StatusInternalServerError, "InternalError", "simulated abort failure")
			return
		}
		f.mu.Lock()
		f.abortRequests = append(f.abortRequests, q.Get("uploadId"))
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)

	case r.Method == http.MethodPut:
		// Single-part PutObject.
		w.Header().Set("ETag", `"single-part-etag"`)
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

type s3XMLError struct {
	XMLName xml.Name `xml:"Error"`
	Code    string   `xml:"Code"`
	Message string   `xml:"Message"`
}

func writeS3Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_ = xml.NewEncoder(w).Encode(s3XMLError{Code: code, Message: message})
}

// TestPut_MultipartOmitsChecksumOnPlaintextEndpoint confirms the plaintext-
// endpoint checksum relaxation (originally added for the single-PutObject
// path in v0.1.4) also reaches manager.Uploader's own, separate checksum
// setting for a multipart upload. Without syncing that setting, every
// UploadPart request would carry a checksum header regardless of endpoint,
// which some S3-compatible backends may not expect or handle correctly.
func TestPut_MultipartOmitsChecksumOnPlaintextEndpoint(t *testing.T) {
	fake := &fakeS3{}
	srv := fake.server() // httptest.Server is plain HTTP
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024)
	if _, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body)); err != nil {
		t.Fatalf("Put: %v", err)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.partHeaders) == 0 {
		t.Fatal("expected at least one UploadPart request")
	}
	for i, h := range fake.partHeaders {
		for name := range h {
			if strings.HasPrefix(strings.ToLower(name), "x-amz-checksum") ||
				strings.EqualFold(name, "X-Amz-Sdk-Checksum-Algorithm") {
				t.Errorf("UploadPart #%d sent checksum header %q against a plaintext endpoint", i+1, name)
			}
		}
	}
}

// TestNew_RejectsUploadPartSizeBelowMinimum confirms an UploadPartSize under
// S3's 5MiB minimum is rejected at construction time with ErrInvalidConfig,
// rather than being silently accepted and only surfacing as an opaque SDK
// error the first time a large Put is attempted.
func TestNew_RejectsUploadPartSizeBelowMinimum(t *testing.T) {
	_, err := New(context.Background(), Config{
		Bucket:         "test-bucket",
		Region:         "us-east-1",
		UploadPartSize: 1024 * 1024, // 1MiB, below manager.MinUploadPartSize (5MiB)
	})
	if !errors.Is(err, objstore.ErrInvalidConfig) {
		t.Errorf("err = %v, want objstore.ErrInvalidConfig", err)
	}
}

// TestNew_AcceptsUploadPartSizeAtMinimum confirms exactly the minimum is
// still valid (only values strictly below it are rejected).
func TestNew_AcceptsUploadPartSizeAtMinimum(t *testing.T) {
	_, err := New(context.Background(), Config{
		Bucket:         "test-bucket",
		Region:         "us-east-1",
		UploadPartSize: 5 * 1024 * 1024,
	})
	if err != nil {
		t.Errorf("New: unexpected error for UploadPartSize at the minimum: %v", err)
	}
}

func newTestStorage(t *testing.T, endpoint string) *Storage {
	t.Helper()
	st, err := New(context.Background(), Config{
		Bucket:          "test-bucket",
		Region:          "us-east-1",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		Endpoint:        endpoint,
		UsePathStyle:    true,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return st
}

// TestPut_MultipartTriggersAboveMinPartSize confirms a body larger than
// manager.MinUploadPartSize (5MiB) goes through CreateMultipartUpload +
// UploadPart + CompleteMultipartUpload rather than a single PutObject.
func TestPut_MultipartTriggersAboveMinPartSize(t *testing.T) {
	fake := &fakeS3{}
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024) // 2 parts: 5MiB + 1MiB
	info, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if info.ETag != "final-etag-2" {
		t.Errorf("ETag = %q, want the multipart-completion ETag", info.ETag)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.uploadID == "" {
		t.Error("expected CreateMultipartUpload to have been called for a 6MiB body")
	}
}

// TestPut_SinglePartForSmallBody confirms a small body still takes the
// direct PutObject path (no CreateMultipartUpload).
func TestPut_SinglePartForSmallBody(t *testing.T) {
	fake := &fakeS3{}
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	info, err := st.Put(context.Background(), "obj.bin", bytes.NewReader([]byte("hello")))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if info.ETag != "single-part-etag" {
		t.Errorf("ETag = %q, want the single-part PutObject ETag", info.ETag)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.uploadID != "" {
		t.Error("expected no CreateMultipartUpload for a 5-byte body")
	}
}

// TestPut_ConditionalOverwriteReachesMultipartCompletion confirms
// WithOverwrite(false) sets If-None-Match on CompleteMultipartUpload, not
// just on the (small-body) PutObject path, since that's where S3 actually
// evaluates the condition for a multipart-created object.
func TestPut_ConditionalOverwriteReachesMultipartCompletion(t *testing.T) {
	fake := &fakeS3{}
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024)
	_, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body), objstore.WithOverwrite(false))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.completeIfNoneMatch != "*" {
		t.Errorf("CompleteMultipartUpload If-None-Match = %q, want %q", fake.completeIfNoneMatch, "*")
	}
}

// TestPut_PreconditionFailedMapsToErrAlreadyExists confirms a conditional
// multipart upload rejected by S3 (existing object) surfaces objstore's
// sentinel error, not a raw SDK error.
func TestPut_PreconditionFailedMapsToErrAlreadyExists(t *testing.T) {
	fake := &fakeS3{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.Method == http.MethodPost && q.Has("uploadId") {
			if r.Body != nil {
				_, _ = io.Copy(io.Discard, r.Body)
				_ = r.Body.Close()
			}
			writeS3Error(w, http.StatusPreconditionFailed, "PreconditionFailed", "object already exists")
			return
		}
		fake.handle(w, r)
	}))
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024)
	_, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body), objstore.WithOverwrite(false))
	if !errors.Is(err, objstore.ErrAlreadyExists) {
		t.Errorf("err = %v, want objstore.ErrAlreadyExists", err)
	}

	// A rejected CompleteMultipartUpload still leaves the upload's parts on
	// S3 (the object was never created), so this needs the same cleanup as
	// any other multipart failure.
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.abortRequests) == 0 {
		t.Error("expected AbortMultipartUpload to still run after a PreconditionFailed completion")
	}
}

// TestPut_AbortsMultipartOnPartFailure confirms an AbortMultipartUpload
// request reaches the server, with the correct upload ID, after a part
// upload fails. It does not isolate abortOrphanedMultipart specifically —
// manager.Uploader's own internal cleanup already satisfies this on an
// ordinary (uncancelled) ctx, so this passes even without that fallback.
// TestPut_RecoversWhenSDKsOwnAbortFails is the test that actually exercises
// abortOrphanedMultipart's reason for existing.
func TestPut_AbortsMultipartOnPartFailure(t *testing.T) {
	fake := &fakeS3{failPartNumber: "2"}
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024) // 2 parts; part 2 always fails
	_, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body))
	if err == nil {
		t.Fatal("expected Put to fail when a part upload fails")
	}
	if errors.Is(err, objstore.ErrAlreadyExists) {
		t.Fatalf("unexpected ErrAlreadyExists for a simulated part failure: %v", err)
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.abortRequests) == 0 {
		t.Fatal("expected at least one AbortMultipartUpload request after a part failure")
	}
	for _, id := range fake.abortRequests {
		if id != fake.uploadID {
			t.Errorf("AbortMultipartUpload uploadId = %q, want %q", id, fake.uploadID)
		}
	}
}

// TestPut_RecoversWhenSDKsOwnAbortFails is the direct regression test for
// abortOrphanedMultipart's actual purpose: manager.Uploader's own abort-on-
// failure reuses the ctx that made the upload fail in the first place (a
// cancellation or deadline), so when that reused ctx is bad, the SDK's own
// abort call fails too. This forces the SDK's internal abort to exhaust its
// retry budget and fail outright, then asserts Put's fallback still lands a
// successful AbortMultipartUpload — i.e. the multipart upload doesn't stay
// orphaned just because the first cleanup attempt failed.
func TestPut_RecoversWhenSDKsOwnAbortFails(t *testing.T) {
	fake := &fakeS3{failPartNumber: "2", failAbortAttempts: 3} // exhausts the default 3-attempt retry budget
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)

	body := bytes.Repeat([]byte("a"), 6*1024*1024) // 2 parts; part 2 always fails
	_, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body))
	if err == nil {
		t.Fatal("expected Put to fail when a part upload fails")
	}

	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.abortAttempts <= fake.failAbortAttempts {
		t.Fatalf("only saw %d AbortMultipartUpload attempt(s), want more than the %d that were made to fail — "+
			"the fallback abort in Put never ran or never retried past the SDK's own failed attempt",
			fake.abortAttempts, fake.failAbortAttempts)
	}
	if len(fake.abortRequests) == 0 {
		t.Fatal("expected the fallback abort to eventually succeed, but no AbortMultipartUpload call did")
	}
	if fake.abortRequests[0] != fake.uploadID {
		t.Errorf("successful AbortMultipartUpload uploadId = %q, want %q", fake.abortRequests[0], fake.uploadID)
	}
}
