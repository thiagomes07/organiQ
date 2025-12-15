// internal/infra/storage/interface.go
package storage

import (
	"context"
	"io"
)

// StorageService define contrato para operações de armazenamento de blobs
type StorageService interface {
	// Upload envia arquivo para storage
	// Retorna URL pública do arquivo ou erro
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (url string, err error)

	// Download baixa arquivo do storage
	// Retorna ReadCloser que deve ser fechado pelo caller
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete remove arquivo do storage
	Delete(ctx context.Context, key string) error

	// Exists verifica se arquivo existe
	Exists(ctx context.Context, key string) (bool, error)

	// GetURL retorna URL pública de um arquivo (sem fazer download)
	GetURL(ctx context.Context, key string) (string, error)

	// HealthCheck verifica se storage está acessível
	HealthCheck(ctx context.Context) error
}

// UploadResult contém metadados do upload
type UploadResult struct {
	Key       string
	URL       string
	Size      int64
	ETag      string
	VersionID string // Para S3 com versioning
}
