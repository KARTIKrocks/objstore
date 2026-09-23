package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	storage "cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/KARTIKrocks/objstore"
)

// fakeGCS serves just enough of the GCS JSON API for Put's precondition
// handling: object metadata reads and simple/multipart uploads.
type fakeGCS struct {
	mu          sync.Mutex
	exists      bool
	uploadFails bool       // respond 412 to the upload
	deleteOnRej bool       // the object is gone once the upload is rejected
	uploadQuery url.Values // query of the last upload request
}

func (f *fakeGCS) handle(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/storage/v1/b/bkt/o/"):
		if !f.exists {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":{"code":404,"message":"No such object"}}`)
			return
		}
		fmt.Fprint(w, `{"bucket":"bkt","name":"obj.txt","etag":"etag-1","generation":"5","metageneration":"2"}`)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/upload/storage/v1/b/bkt/o"):
		f.uploadQuery = r.URL.Query()
		if f.uploadFails {
			if f.deleteOnRej {
				f.exists = false
			}
			w.WriteHeader(http.StatusPreconditionFailed)
			fmt.Fprint(w, `{"error":{"code":412,"message":"Precondition Failed"}}`)
			return
		}
		fmt.Fprint(w, `{"bucket":"bkt","name":"obj.txt","etag":"etag-2","generation":"6","metageneration":"1","updated":"2026-01-02T03:04:05Z"}`)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}

func newFakeStorage(t *testing.T, fake *fakeGCS) *Storage {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(srv.Close)

	client, err := storage.NewClient(context.Background(),
		option.WithEndpoint(srv.URL+"/storage/v1/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return &Storage{client: client, bucket: client.Bucket("bkt"), config: Config{Bucket: "bkt"}}
}

func TestPut_ReturnsETag(t *testing.T) {
	st := newFakeStorage(t, &fakeGCS{})
	info, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if info.ETag != "etag-2" {
		t.Errorf("ETag = %q, want %q", info.ETag, "etag-2")
	}
}

func TestPut_NoOverwriteSendsDoesNotExist(t *testing.T) {
	fake := &fakeGCS{}
	st := newFakeStorage(t, fake)
	if _, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), objstore.WithOverwrite(false)); err != nil {
		t.Fatalf("Put: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if got := fake.uploadQuery.Get("ifGenerationMatch"); got != "0" {
		t.Errorf("ifGenerationMatch = %q, want %q", got, "0")
	}
}

func TestPut_NoOverwriteRejectedMapsToErrAlreadyExists(t *testing.T) {
	st := newFakeStorage(t, &fakeGCS{uploadFails: true})
	_, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), objstore.WithOverwrite(false))
	if !errors.Is(err, objstore.ErrAlreadyExists) {
		t.Errorf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestPut_IfMatchPinsGeneration(t *testing.T) {
	fake := &fakeGCS{exists: true}
	st := newFakeStorage(t, fake)
	if _, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), objstore.WithIfMatch("etag-1")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if got := fake.uploadQuery.Get("ifGenerationMatch"); got != "5" {
		t.Errorf("ifGenerationMatch = %q, want %q", got, "5")
	}
	if got := fake.uploadQuery.Get("ifMetagenerationMatch"); got != "2" {
		t.Errorf("ifMetagenerationMatch = %q, want %q", got, "2")
	}
}

func TestPut_IfMatchErrors(t *testing.T) {
	tests := []struct {
		name         string
		fake         *fakeGCS
		etag         string
		wantNotFound bool
	}{
		{"stale etag", &fakeGCS{exists: true}, "etag-0", false},
		{"missing object", &fakeGCS{}, "etag-1", true},
		{"changed after read", &fakeGCS{exists: true, uploadFails: true}, "etag-1", false},
		{"deleted after read", &fakeGCS{exists: true, uploadFails: true, deleteOnRej: true}, "etag-1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := newFakeStorage(t, tt.fake)
			_, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), objstore.WithIfMatch(tt.etag))
			if !errors.Is(err, objstore.ErrPreconditionFailed) {
				t.Errorf("err = %v, want ErrPreconditionFailed", err)
			}
			if got := errors.Is(err, objstore.ErrNotFound); got != tt.wantNotFound {
				t.Errorf("errors.Is(err, ErrNotFound) = %v, want %v", got, tt.wantNotFound)
			}
		})
	}
}
