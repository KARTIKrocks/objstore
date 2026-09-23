package objstore

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/iotest"
	"time"
	"uuid"
)

// conditionalBackends runs each conditional-write test against every
// in-module backend, since they share one contract.
func conditionalBackends(t *testing.T) map[string]Storage {
	return map[string]Storage{
		"local":  newTestLocalStorage(t),
		"memory": NewMemoryStorage(),
	}
}

func TestConditionalPut_ETagRoundTrips(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			put, err := store.Put(ctx, "a/doc.txt", strings.NewReader("v1"))
			if err != nil {
				t.Fatalf("Put: %v", err)
			}
			if put.ETag == "" {
				t.Fatal("Put returned an empty ETag")
			}

			stat, err := store.Stat(ctx, "a/doc.txt")
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			if stat.ETag != put.ETag {
				t.Errorf("Stat ETag = %q, want Put's %q", stat.ETag, put.ETag)
			}

			list, err := store.List(ctx, "a/")
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(list.Files) != 1 || list.Files[0].ETag != put.ETag {
				t.Errorf("List ETags = %+v, want one file with %q", list.Files, put.ETag)
			}
		})
	}
}

func TestConditionalPut_IfMatch(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			v1, err := store.Put(ctx, "doc.txt", strings.NewReader("v1"))
			if err != nil {
				t.Fatalf("Put v1: %v", err)
			}

			// Same size as v1, written immediately: a coarse mtime must not
			// let the ETag stay the same.
			v2, err := store.Put(ctx, "doc.txt", strings.NewReader("v2"), WithIfMatch(v1.ETag))
			if err != nil {
				t.Fatalf("Put with current ETag: %v", err)
			}
			if v2.ETag == v1.ETag {
				t.Fatalf("ETag unchanged after rewrite: %q", v2.ETag)
			}

			_, err = store.Put(ctx, "doc.txt", strings.NewReader("v3"), WithIfMatch(v1.ETag))
			if !errors.Is(err, ErrPreconditionFailed) {
				t.Fatalf("Put with stale ETag: err = %v, want ErrPreconditionFailed", err)
			}
			if got, _ := GetString(ctx, store, "doc.txt"); got != "v2" {
				t.Errorf("content after rejected Put = %q, want %q", got, "v2")
			}
		})
	}
}

// TestLocalPut_ModTimeAlwaysAdvances pins the guarantee behind local ETags:
// a rewrite never keeps (or goes behind) the previous mtime. The previous
// version is dated in the future to stand in for a coarse clock that has not
// ticked since the last write.
func TestLocalPut_ModTimeAlwaysAdvances(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)

	if _, err := store.Put(ctx, "doc.txt", strings.NewReader("v1")); err != nil {
		t.Fatalf("Put v1: %v", err)
	}
	fullPath, _ := store.fullPath("doc.txt")
	future := time.Now().Add(time.Hour).Truncate(time.Second)
	if err := os.Chtimes(fullPath, time.Time{}, future); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	stale, err := store.Stat(ctx, "doc.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	v2, err := store.Put(ctx, "doc.txt", strings.NewReader("v2"), WithIfMatch(stale.ETag))
	if err != nil {
		t.Fatalf("Put v2: %v", err)
	}
	if !v2.LastModified.After(future) {
		t.Errorf("LastModified = %v, want after previous version's %v", v2.LastModified, future)
	}
	if v2.ETag == stale.ETag {
		t.Errorf("ETag unchanged after rewrite: %q", v2.ETag)
	}
}

// TestLocalPut_WaitHonorsContext confirms a Put queued behind another Put to
// the same path gives up when its context ends, rather than waiting for the
// first upload to finish.
func TestLocalPut_WaitHonorsContext(t *testing.T) {
	store := newTestLocalStorage(t)

	body, feed := io.Pipe()
	firstDone := make(chan error, 1)
	go func() {
		_, err := store.Put(context.Background(), "doc.txt", body)
		firstDone <- err
	}()
	if _, err := feed.Write([]byte("partial")); err != nil { // first Put now holds the lock
		t.Fatalf("Write: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := store.Put(ctx, "doc.txt", strings.NewReader("x"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("queued Put err = %v, want context.DeadlineExceeded", err)
	}

	_ = feed.Close()
	if err := <-firstDone; err != nil {
		t.Errorf("first Put: %v", err)
	}
}

// TestLocalPut_FailedOverwriteKeepsPreviousVersion confirms a Put whose body
// fails part-way leaves the committed file untouched and no temp file behind,
// matching cloud backends, where a failed upload never replaces the object.
func TestLocalPut_FailedOverwriteKeepsPreviousVersion(t *testing.T) {
	ctx := context.Background()
	for name, opts := range map[string]func(etag string) []PutOption{
		"overwrite": func(string) []PutOption { return nil },
		"if match":  func(etag string) []PutOption { return []PutOption{WithIfMatch(etag)} },
	} {
		t.Run(name, func(t *testing.T) {
			store := newTestLocalStorage(t)
			v1, err := store.Put(ctx, "doc.txt", strings.NewReader("v1"))
			if err != nil {
				t.Fatalf("Put v1: %v", err)
			}

			failing := io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(errors.New("boom")))
			if _, err := store.Put(ctx, "doc.txt", failing, opts(v1.ETag)...); err == nil {
				t.Fatal("Put with failing body succeeded")
			}

			if got, _ := GetString(ctx, store, "doc.txt"); got != "v1" {
				t.Errorf("content after failed Put = %q, want %q", got, "v1")
			}
			if stat, _ := store.Stat(ctx, "doc.txt"); stat == nil || stat.ETag != v1.ETag {
				t.Errorf("ETag changed after failed Put: %+v, want %q", stat, v1.ETag)
			}
			entries, err := os.ReadDir(store.config.BasePath)
			if err != nil {
				t.Fatalf("ReadDir: %v", err)
			}
			if len(entries) != 1 {
				t.Errorf("directory has %d entries, want only doc.txt (temp file leaked?)", len(entries))
			}
		})
	}
}

// TestLocalPut_IfMatchStatErrorIsNotNotFound confirms an existing path that
// can't be stat-ed is reported as a permission error, not as a deleted object.
func TestLocalPut_IfMatchStatErrorIsNotNotFound(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs POSIX permissions enforced for the current user")
	}
	ctx := context.Background()
	store := newTestLocalStorage(t)
	if _, err := store.Put(ctx, "locked/doc.txt", strings.NewReader("v1")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	dir := filepath.Join(store.config.BasePath, "locked")
	if err := os.Chmod(dir, 0); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	_, err := store.Put(ctx, "locked/doc.txt", strings.NewReader("v2"), WithIfMatch("anything"))
	if !errors.Is(err, ErrPermission) {
		t.Errorf("err = %v, want ErrPermission", err)
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrPreconditionFailed) {
		t.Errorf("err = %v, must not claim the object is missing or changed", err)
	}
}

// TestLocalList_SkipsOnlyInternalTempFiles confirms a Put's temp file is
// never listed as an object, while an ordinary object that merely shares the
// temp prefix still is, as it would be on any other backend.
func TestLocalList_SkipsOnlyInternalTempFiles(t *testing.T) {
	ctx := context.Background()
	store := newTestLocalStorage(t)
	for _, key := range []string{"doc.txt", tempFilePrefix + "report.txt"} {
		if _, err := store.Put(ctx, key, strings.NewReader("v1")); err != nil {
			t.Fatalf("Put %s: %v", key, err)
		}
	}
	tmp := filepath.Join(store.config.BasePath, tempFilePrefix+uuid.New().String())
	if err := os.WriteFile(tmp, []byte("partial"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	list, err := store.List(ctx, "", WithRecursive(true))
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := make([]string, 0, len(list.Files))
	for _, f := range list.Files {
		got = append(got, f.Path)
	}
	want := []string{tempFilePrefix + "report.txt", "doc.txt"}
	if !slices.Equal(slices.Sorted(slices.Values(got)), slices.Sorted(slices.Values(want))) {
		t.Errorf("List = %v, want %v", got, want)
	}
}

// TestLocalPut_OverwriteKeepsFileMode confirms replacing a file (which swaps
// in a new inode) keeps its permissions instead of resetting them to
// FilePermissions — overwriting a 0600 file must not make it world-readable.
func TestLocalPut_OverwriteKeepsFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits")
	}
	ctx := context.Background()
	store := newTestLocalStorage(t)
	if _, err := store.Put(ctx, "secret.txt", strings.NewReader("v1")); err != nil {
		t.Fatalf("Put v1: %v", err)
	}
	fullPath, _ := store.fullPath("secret.txt")
	if err := os.Chmod(fullPath, 0o600); err != nil {
		t.Fatalf("Chmod: %v", err)
	}

	if _, err := store.Put(ctx, "secret.txt", strings.NewReader("v2")); err != nil {
		t.Fatalf("Put v2: %v", err)
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("mode after overwrite = %v, want %v", got, os.FileMode(0o600))
	}
}

func TestConditionalPut_IfMatchMissingObject(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			_, err := store.Put(ctx, "missing.txt", strings.NewReader("x"), WithIfMatch("anything"))
			if !errors.Is(err, ErrPreconditionFailed) || !errors.Is(err, ErrNotFound) {
				t.Fatalf("err = %v, want both ErrPreconditionFailed and ErrNotFound", err)
			}
			if ok, _ := store.Exists(ctx, "missing.txt"); ok {
				t.Error("rejected Put created the object")
			}
		})
	}
}

func TestConditionalPut_IfMatchIgnoresOverwrite(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			v1, err := store.Put(ctx, "doc.txt", strings.NewReader("v1"))
			if err != nil {
				t.Fatalf("Put: %v", err)
			}
			if _, err := store.Put(ctx, "doc.txt", strings.NewReader("v2"), WithOverwrite(false), WithIfMatch(v1.ETag)); err != nil {
				t.Fatalf("Put: %v, want WithIfMatch to take precedence", err)
			}
		})
	}
}

// TestConditionalPut_ConcurrentCompareAndSwap races writers that all hold the
// same ETag: exactly one may win.
func TestConditionalPut_ConcurrentCompareAndSwap(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			v1, err := store.Put(ctx, "counter", strings.NewReader("0"))
			if err != nil {
				t.Fatalf("Put: %v", err)
			}

			const writers = 16
			var wg sync.WaitGroup
			errs := make(chan error, writers)
			for range writers {
				wg.Go(func() {
					_, err := store.Put(ctx, "counter", strings.NewReader("1"), WithIfMatch(v1.ETag))
					errs <- err
				})
			}
			wg.Wait()
			close(errs)

			wins := 0
			for err := range errs {
				switch {
				case err == nil:
					wins++
				case !errors.Is(err, ErrPreconditionFailed):
					t.Errorf("unexpected error: %v", err)
				}
			}
			if wins != 1 {
				t.Errorf("%d writers succeeded, want exactly 1", wins)
			}
		})
	}
}

func TestConditionalPut_NoOverwriteIsAtomic(t *testing.T) {
	ctx := context.Background()
	for name, store := range conditionalBackends(t) {
		t.Run(name, func(t *testing.T) {
			const writers = 16
			var wg sync.WaitGroup
			errs := make(chan error, writers)
			for range writers {
				wg.Go(func() {
					_, err := store.Put(ctx, "once.txt", strings.NewReader("x"), WithOverwrite(false))
					errs <- err
				})
			}
			wg.Wait()
			close(errs)

			wins := 0
			for err := range errs {
				switch {
				case err == nil:
					wins++
				case !errors.Is(err, ErrAlreadyExists):
					t.Errorf("unexpected error: %v", err)
				}
			}
			if wins != 1 {
				t.Errorf("%d writers succeeded, want exactly 1", wins)
			}
		})
	}
}
