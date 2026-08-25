package s3

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	completeIfNoneMatch string
	abortRequests       []string // uploadIds seen on AbortMultipartUpload
}

func (f *fakeS3) server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f.handle))
}

func (f *fakeS3) handle(w http.ResponseWriter, r *http.Request) {
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
}

// TestPut_AbortsMultipartOnPartFailure confirms that when a part upload
// fails partway through, Put's abortOrphanedMultipart cleanup reaches the
// server with the same upload ID CreateMultipartUpload handed out — so a
// failed large upload doesn't leave billable orphaned parts on S3.
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
