package storage

import (
	"net/http"
	"path/filepath"
	"strings"
)

func detectContentType(key string, data []byte) string {
	if len(data) > 512 {
		return http.DetectContentType(data[:512])
	}
	if len(data) > 0 {
		return http.DetectContentType(data)
	}

	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".csv":
		return "text/csv"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
