package azure

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/KARTIKrocks/objstore"
)

// fakeAzure records the precondition headers on blob writes and optionally
// rejects them with the given status and x-ms-error-code.
type fakeAzure struct {
	mu          sync.Mutex
	ifMatch     string
	ifNoneMatch string
	failStatus  int
	failCode    string
}

func (f *fakeAzure) handle(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	f.ifMatch = r.Header.Get("If-Match")
	f.ifNoneMatch = r.Header.Get("If-None-Match")
	if f.failStatus != 0 {
		w.Header().Set("x-ms-error-code", f.failCode)
		w.WriteHeader(f.failStatus)
		return
	}
	w.Header().Set("ETag", `"0x8DNEW"`)
	w.Header().Set("Last-Modified", "Fri, 02 Jan 2026 03:04:05 GMT")
	w.WriteHeader(http.StatusCreated)
}

func newFakeStorage(t *testing.T, fake *fakeAzure) *Storage {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(srv.Close)

	client, err := azblob.NewClientWithNoCredential(srv.URL+"/", &azblob.ClientOptions{
		ClientOptions: policy.ClientOptions{Retry: policy.RetryOptions{MaxRetries: -1}},
	})
	if err != nil {
		t.Fatalf("NewClientWithNoCredential: %v", err)
	}
	return &Storage{client: client, config: Config{ContainerName: "ctr"}}
}

func TestPut_ReturnsETag(t *testing.T) {
	st := newFakeStorage(t, &fakeAzure{})
	info, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"))
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if info.ETag != `"0x8DNEW"` {
		t.Errorf("ETag = %q, want %q", info.ETag, `"0x8DNEW"`)
	}
}

func TestPut_ConditionHeaders(t *testing.T) {
	tests := []struct {
		name            string
		opts            []objstore.PutOption
		wantIfMatch     string
		wantIfNoneMatch string
	}{
		{"unconditional", nil, "", ""},
		{"no overwrite", []objstore.PutOption{objstore.WithOverwrite(false)}, "", "*"},
		{"if match", []objstore.PutOption{objstore.WithIfMatch(`"0x8DOLD"`)}, `"0x8DOLD"`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeAzure{}
			st := newFakeStorage(t, fake)
			if _, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), tt.opts...); err != nil {
				t.Fatalf("Put: %v", err)
			}
			fake.mu.Lock()
			defer fake.mu.Unlock()
			if fake.ifMatch != tt.wantIfMatch {
				t.Errorf("If-Match = %q, want %q", fake.ifMatch, tt.wantIfMatch)
			}
			if fake.ifNoneMatch != tt.wantIfNoneMatch {
				t.Errorf("If-None-Match = %q, want %q", fake.ifNoneMatch, tt.wantIfNoneMatch)
			}
		})
	}
}

func TestPut_ConditionErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		opts   []objstore.PutOption
		status int
		code   string
		want   []error
		notErr error
	}{
		{"no overwrite, exists (409)", []objstore.PutOption{objstore.WithOverwrite(false)}, http.StatusConflict, "BlobAlreadyExists",
			[]error{objstore.ErrAlreadyExists}, objstore.ErrPreconditionFailed},
		{"no overwrite, exists (412)", []objstore.PutOption{objstore.WithOverwrite(false)}, http.StatusPreconditionFailed, "ConditionNotMet",
			[]error{objstore.ErrAlreadyExists}, objstore.ErrPreconditionFailed},
		{"if match, stale", []objstore.PutOption{objstore.WithIfMatch(`"0x8DOLD"`)}, http.StatusPreconditionFailed, "ConditionNotMet",
			[]error{objstore.ErrPreconditionFailed}, objstore.ErrAlreadyExists},
		{"if match, missing", []objstore.PutOption{objstore.WithIfMatch(`"0x8DOLD"`)}, http.StatusNotFound, "BlobNotFound",
			[]error{objstore.ErrPreconditionFailed, objstore.ErrNotFound}, objstore.ErrAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := newFakeStorage(t, &fakeAzure{failStatus: tt.status, failCode: tt.code})
			_, err := st.Put(context.Background(), "obj.txt", strings.NewReader("hi"), tt.opts...)
			for _, want := range tt.want {
				if !errors.Is(err, want) {
					t.Errorf("err = %v, want it to match %v", err, want)
				}
			}
			if errors.Is(err, tt.notErr) {
				t.Errorf("err = %v, must not match %v", err, tt.notErr)
			}
		})
	}
}
