package objstore

import (
	"path/filepath"
	"strings"
)

// mimeTypes maps file extensions to MIME types.
var mimeTypes = map[string]string{
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".md":   "text/markdown",

	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".ico":  "image/x-icon",
	".bmp":  "image/bmp",

	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".m4a":  "audio/mp4",
	".flac": "audio/flac",

	".mp4":  "video/mp4",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".mkv":  "video/x-matroska",

	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

	".zip": "application/zip",
	".tar": "application/x-tar",
	".gz":  "application/gzip",
	".rar": "application/vnd.rar",
	".7z":  "application/x-7z-compressed",

	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
	".eot":   "application/vnd.ms-fontobject",
}

// DetectContentType detects the MIME type from file extension.
func DetectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// NormalizePath normalizes a path for consistent storage.
func NormalizePath(path string) string {
	path = strings.TrimPrefix(path, "/")
	return filepath.ToSlash(filepath.Clean(path))
}
