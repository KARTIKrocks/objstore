package s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/KARTIKrocks/objstore"
)

func TestPut_IfMatchSendsQuotedETag(t *testing.T) {
	var (
		mu  sync.Mutex
		got string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		mu.Lock()
		got = r.Header.Get("If-Match")
		mu.Unlock()
		w.Header().Set("ETag", `"new-etag"`)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	st := newTestStorage(t, srv.URL)
	info, err := st.Put(context.Background(), "obj.txt", bytes.NewReader([]byte("hi")), objstore.WithIfMatch("old-etag"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if info.ETag != "new-etag" {
		t.Errorf("ETag = %q, want %q", info.ETag, "new-etag")
	}
	mu.Lock()
	defer mu.Unlock()
	if got != `"old-etag"` {
		t.Errorf("If-Match = %q, want %q", got, `"old-etag"`)
	}
}

// TestPut_IfMatchReachesMultipartCompletion confirms If-Match is sent where
// S3 evaluates it for a multipart-created object.
func TestPut_IfMatchReachesMultipartCompletion(t *testing.T) {
	fake := &fakeS3{}
	srv := fake.server()
	defer srv.Close()

	st := newTestStorage(t, srv.URL)
	body := bytes.Repeat([]byte("a"), 6*1024*1024)
	if _, err := st.Put(context.Background(), "obj.bin", bytes.NewReader(body), objstore.WithIfMatch("old-etag")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.completeIfMatch != `"old-etag"` {
		t.Errorf("CompleteMultipartUpload If-Match = %q, want %q", fake.completeIfMatch, `"old-etag"`)
	}
	if fake.completeIfNoneMatch != "" {
		t.Errorf("CompleteMultipartUpload If-None-Match = %q, want none", fake.completeIfNoneMatch)
	}
}

func TestPut_IfMatchErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		code       string
		wantNotFnd bool
	}{
		{"stale etag", http.StatusPreconditionFailed, "PreconditionFailed", false},
		{"missing object", http.StatusNotFound, "NoSuchKey", true},
		{"concurrent write conflict", http.StatusConflict, "ConditionalRequestConflict", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				writeS3Error(w, tt.status, tt.code, "condition not met")
			}))
			defer srv.Close()

			st := newTestStorage(t, srv.URL)
			_, err := st.Put(context.Background(), "obj.txt", bytes.NewReader([]byte("hi")), objstore.WithIfMatch("old-etag"))
			if !errors.Is(err, objstore.ErrPreconditionFailed) {
				t.Errorf("err = %v, want ErrPreconditionFailed", err)
			}
			if errors.Is(err, objstore.ErrAlreadyExists) {
				t.Errorf("err = %v, must not be ErrAlreadyExists", err)
			}
			if got := errors.Is(err, objstore.ErrNotFound); got != tt.wantNotFnd {
				t.Errorf("errors.Is(err, ErrNotFound) = %v, want %v", got, tt.wantNotFnd)
			}
		})
	}
}
