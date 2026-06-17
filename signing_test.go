package objstore

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"
)

const testSecret = "test-signing-secret"

func TestVerifySignedURL_RoundTrip(t *testing.T) {
	raw, err := buildSignedURL("https://cdn.example.com", "uploads/doc.pdf", testSecret,
		"PUT", "application/pdf", time.Now().Add(5*time.Minute))
	if err != nil {
		t.Fatalf("buildSignedURL: %v", err)
	}

	got, err := VerifySignedURL(raw, testSecret)
	if err != nil {
		t.Fatalf("VerifySignedURL: %v", err)
	}
	if got.Method != "PUT" {
		t.Errorf("method = %q, want PUT", got.Method)
	}
	if got.ContentType != "application/pdf" {
		t.Errorf("content type = %q, want application/pdf", got.ContentType)
	}
	if got.Path != "/uploads/doc.pdf" {
		t.Errorf("path = %q, want /uploads/doc.pdf", got.Path)
	}
}

func TestVerifySignedURL_TamperedSignature(t *testing.T) {
	raw, _ := buildSignedURL("https://cdn.example.com", "a.txt", testSecret,
		"GET", "", time.Now().Add(time.Minute))

	// Flip the signed method without re-signing.
	tampered := strings.Replace(raw, "X-Obj-Method=GET", "X-Obj-Method=PUT", 1)
	if _, err := VerifySignedURL(tampered, testSecret); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatalf("err = %v, want ErrSignatureInvalid", err)
	}
}

func TestVerifySignedURL_WrongSecret(t *testing.T) {
	raw, _ := buildSignedURL("https://cdn.example.com", "a.txt", testSecret,
		"GET", "", time.Now().Add(time.Minute))
	if _, err := VerifySignedURL(raw, "other-secret"); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatalf("err = %v, want ErrSignatureInvalid", err)
	}
}

func TestVerifySignedURL_MissingExpiry(t *testing.T) {
	raw, _ := buildSignedURL("https://cdn.example.com", "a.txt", testSecret,
		"GET", "", time.Now().Add(time.Minute))

	// Drop the expiry parameter, leaving the signature in place.
	u, _ := url.Parse(raw)
	q := u.Query()
	q.Del(signParamExpires)
	u.RawQuery = q.Encode()

	if _, err := VerifySignedURL(u.String(), testSecret); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatalf("err = %v, want ErrSignatureInvalid", err)
	}
}

func TestVerifySignedURL_Expired(t *testing.T) {
	raw, _ := buildSignedURL("https://cdn.example.com", "a.txt", testSecret,
		"GET", "", time.Now().Add(-time.Second))
	if _, err := VerifySignedURL(raw, testSecret); !errors.Is(err, ErrSignatureExpired) {
		t.Fatalf("err = %v, want ErrSignatureExpired", err)
	}
}

func TestVerifySignedURL_MissingSignature(t *testing.T) {
	if _, err := VerifySignedURL("https://cdn.example.com/a.txt", testSecret); !errors.Is(err, ErrSignatureInvalid) {
		t.Fatalf("err = %v, want ErrSignatureInvalid", err)
	}
}
