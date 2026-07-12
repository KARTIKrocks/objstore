package s3

import "testing"

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
