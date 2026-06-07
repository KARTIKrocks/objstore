package objstore

import (
	"testing"
	"time"
)

func TestApplyPutOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		opts := ApplyPutOptions(nil)
		if !opts.Overwrite {
			t.Error("expected Overwrite to default to true")
		}
		if opts.ContentType != "" {
			t.Errorf("expected empty ContentType, got %q", opts.ContentType)
		}
	})

	t.Run("with all options", func(t *testing.T) {
		meta := map[string]string{"key": "value"}
		opts := ApplyPutOptions([]PutOption{
			WithContentType("image/png"),
			WithMetadata(meta),
			WithCacheControl("max-age=3600"),
			WithACL("public-read"),
			WithOverwrite(false),
		})

		if opts.ContentType != "image/png" {
			t.Errorf("ContentType = %q, want %q", opts.ContentType, "image/png")
		}
		if opts.Metadata["key"] != "value" {
			t.Error("Metadata not set correctly")
		}
		if opts.CacheControl != "max-age=3600" {
			t.Errorf("CacheControl = %q, want %q", opts.CacheControl, "max-age=3600")
		}
		if opts.ACL != "public-read" {
			t.Errorf("ACL = %q, want %q", opts.ACL, "public-read")
		}
		if opts.Overwrite {
			t.Error("expected Overwrite to be false")
		}
	})
}

func TestApplyListOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		opts := ApplyListOptions(nil)
		if opts.MaxKeys != 1000 {
			t.Errorf("MaxKeys = %d, want 1000", opts.MaxKeys)
		}
		if opts.Delimiter != "/" {
			t.Errorf("Delimiter = %q, want %q", opts.Delimiter, "/")
		}
		if opts.Recursive {
			t.Error("expected Recursive to default to false")
		}
	})

	t.Run("with options", func(t *testing.T) {
		opts := ApplyListOptions([]ListOption{
			WithMaxKeys(50),
			WithDelimiter("|"),
			WithToken("abc"),
			WithRecursive(true),
		})

		if opts.MaxKeys != 50 {
			t.Errorf("MaxKeys = %d, want 50", opts.MaxKeys)
		}
		if opts.Delimiter != "|" {
			t.Errorf("Delimiter = %q, want %q", opts.Delimiter, "|")
		}
		if opts.Token != "abc" {
			t.Errorf("Token = %q, want %q", opts.Token, "abc")
		}
		if !opts.Recursive {
			t.Error("expected Recursive to be true")
		}
	})
}

func TestApplySignedURLOptions(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		opts := ApplySignedURLOptions(nil)
		if opts.Expires != 15*time.Minute {
			t.Errorf("Expires = %v, want 15m", opts.Expires)
		}
		if opts.Method != "GET" {
			t.Errorf("Method = %q, want %q", opts.Method, "GET")
		}
	})

	t.Run("with options", func(t *testing.T) {
		headers := map[string]string{"X-Custom": "val"}
		opts := ApplySignedURLOptions([]SignedURLOption{
			WithExpires(1 * time.Hour),
			WithMethod("PUT"),
			WithSignedContentType("application/json"),
			WithHeaders(headers),
		})

		if opts.Expires != 1*time.Hour {
			t.Errorf("Expires = %v, want 1h", opts.Expires)
		}
		if opts.Method != "PUT" {
			t.Errorf("Method = %q, want %q", opts.Method, "PUT")
		}
		if opts.ContentType != "application/json" {
			t.Errorf("ContentType = %q, want %q", opts.ContentType, "application/json")
		}
		if opts.Headers["X-Custom"] != "val" {
			t.Error("Headers not set correctly")
		}
	})
}
