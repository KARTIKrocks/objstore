package objstore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Signed-URL query parameters carried on URLs produced by the local and memory
// backends. They are namespaced so they don't collide with an application's own
// query parameters.
const (
	signParamExpires     = "X-Obj-Expires"      // expiry as Unix seconds
	signParamMethod      = "X-Obj-Method"       // HTTP method the URL authorizes
	signParamContentType = "X-Obj-Content-Type" // required Content-Type for PUT, if set
	signParamSignature   = "X-Obj-Signature"    // hex HMAC-SHA256 of the canonical string
)

// SignedRequest is the verified content of a signed URL: what operation it
// authorizes, on which object path, until when. Returned by VerifySignedURL.
type SignedRequest struct {
	Path        string    // URL path the signature covers (includes any BaseURL path prefix)
	Method      string    // authorized HTTP method (GET, PUT, …)
	ContentType string    // Content-Type the upload must use, empty if unconstrained
	Expires     time.Time // instant after which the URL is invalid
}

// buildSignedURL produces an HMAC-signed URL for objectPath under baseURL,
// authorizing method until expires (with an optional required content type for
// uploads). The local and memory backends use it; cloud backends delegate to
// their provider's native presigning instead.
func buildSignedURL(baseURL, objectPath, secret, method, contentType string, expires time.Time) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base URL: %v", ErrInvalidConfig, err)
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(objectPath, "/")

	q := u.Query()
	q.Set(signParamExpires, strconv.FormatInt(expires.Unix(), 10))
	q.Set(signParamMethod, method)
	if contentType != "" {
		q.Set(signParamContentType, contentType)
	}
	q.Set(signParamSignature, sign(secret, method, u.Path, expires.Unix(), contentType))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// VerifySignedURL validates a URL produced by buildSignedURL against secret and
// returns what it authorizes. It returns ErrSignatureInvalid if the signature is
// missing, malformed, or does not match, and ErrSignatureExpired if the URL has
// expired. Applications call this from the HTTP handler that serves local/memory
// signed URLs; cloud backends do not use it (the provider verifies natively).
func VerifySignedURL(rawURL, secret string) (*SignedRequest, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignatureInvalid, err)
	}
	q := u.Query()

	got := q.Get(signParamSignature)
	if got == "" {
		return nil, fmt.Errorf("%w: missing signature", ErrSignatureInvalid)
	}
	expiresRaw := q.Get(signParamExpires)
	if expiresRaw == "" {
		return nil, fmt.Errorf("%w: missing expiry", ErrSignatureInvalid)
	}
	expiresUnix, err := strconv.ParseInt(expiresRaw, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed expiry", ErrSignatureInvalid)
	}
	method := q.Get(signParamMethod)
	contentType := q.Get(signParamContentType)

	want := sign(secret, method, u.Path, expiresUnix, contentType)
	if !hmac.Equal([]byte(got), []byte(want)) {
		return nil, ErrSignatureInvalid
	}
	if time.Now().After(time.Unix(expiresUnix, 0)) {
		return nil, ErrSignatureExpired
	}
	return &SignedRequest{
		Path:        u.Path,
		Method:      method,
		ContentType: contentType,
		Expires:     time.Unix(expiresUnix, 0),
	}, nil
}

// sign computes the hex HMAC-SHA256 over the canonical fields. The newline-joined
// layout is unambiguous because none of the fields can contain a newline.
func sign(secret, method, path string, expiresUnix int64, contentType string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		method,
		path,
		strconv.FormatInt(expiresUnix, 10),
		contentType,
	}, "\n")))
	return hex.EncodeToString(mac.Sum(nil))
}
