package objstore

import "testing"

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"photo.jpg", "image/jpeg"},
		{"photo.JPEG", "image/jpeg"},
		{"style.css", "text/css"},
		{"script.js", "application/javascript"},
		{"data.json", "application/json"},
		{"page.html", "text/html"},
		{"page.htm", "text/html"},
		{"doc.pdf", "application/pdf"},
		{"song.mp3", "audio/mpeg"},
		{"video.mp4", "video/mp4"},
		{"archive.zip", "application/zip"},
		{"font.woff2", "font/woff2"},
		{"image.webp", "image/webp"},
		{"image.svg", "image/svg+xml"},
		{"notes.md", "text/markdown"},
		{"data.csv", "text/csv"},
		{"doc.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"sheet.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		// Path with directories
		{"uploads/2024/photo.png", "image/png"},
		// Unknown extension
		{"file.xyz", "application/octet-stream"},
		// No extension
		{"Makefile", "application/octet-stream"},
		// Empty path
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := DetectContentType(tt.path)
			if got != tt.want {
				t.Errorf("DetectContentType(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/foo/bar", "foo/bar"},
		{"foo/bar", "foo/bar"},
		{"foo//bar", "foo/bar"},
		{"/foo/../bar", "bar"},
		{"./foo/bar", "foo/bar"},
		{"foo/./bar", "foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizePath(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
