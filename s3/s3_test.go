package s3

import (
	"math"
	"testing"

	"github.com/KARTIKrocks/objstore"
)

// TestIsPlaintextEndpoint pins which endpoints get the relaxed request-checksum
// mode. Only plain-HTTP endpoints do: over HTTP the SDK cannot send the trailing
// checksum it computes for streaming bodies, so uploads fail outright. TLS
// endpoints — including an empty one, which means AWS — keep the strict default.
func TestIsPlaintextEndpoint(t *testing.T) {
	tests := []struct {
		endpoint string
		want     bool
	}{
		{"", false}, // AWS S3
		{"https://nyc3.digitaloceanspaces.com", false},
		{"https://localhost:9000", false},
		{"http://localhost:9000", true}, // MinIO in local dev
		{"http://minio:9000", true},
		{"HTTP://LOCALHOST:9000", true}, // scheme is case-insensitive
	}
	for _, tt := range tests {
		if got := isPlaintextEndpoint(tt.endpoint); got != tt.want {
			t.Errorf("isPlaintextEndpoint(%q) = %v, want %v", tt.endpoint, got, tt.want)
		}
	}
}

// TestFormatRange pins the HTTP Range header S3 receives for each GetOptions
// shape, including the sentinel values (Offset==0 and non-positive Length)
// that mean "no range requested".
func TestFormatRange(t *testing.T) {
	tests := []struct {
		name   string
		offset int64
		length int64
		want   string
	}{
		{"no range", 0, 0, ""},
		{"no range, negative length", 0, -1, ""},
		{"bounded range", 3, 4, "bytes=3-6"},
		{"offset to end, zero length", 7, 0, "bytes=7-"},
		{"offset to end, negative length", 7, -1, "bytes=7-"},
		{"single byte", 0, 1, "bytes=0-0"},
		{"length overflows int64, falls back to open-ended", 100, math.MaxInt64, "bytes=100-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := objstore.ApplyGetOptions([]objstore.GetOption{objstore.WithRange(tt.offset, tt.length)})
			if got := formatRange(options); got != tt.want {
				t.Errorf("formatRange(offset=%d, length=%d) = %q, want %q", tt.offset, tt.length, got, tt.want)
			}
		})
	}
}
