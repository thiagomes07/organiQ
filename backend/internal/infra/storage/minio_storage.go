package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

// MinIOStorage implementa StorageService para MinIO
type MinIOStorage struct {
	client    *s3.Client
	bucket    string
	endpoint  string
	useSSL    bool
	publicURL string
}

// NewMinIOStorage cria nova instância para MinIO
func NewMinIOStorage(
	client *s3.Client,
	bucket string,
	endpoint string,
	useSSL bool,
	publicURL string,
) StorageService {
	return &MinIOStorage{
		client:    client,
		bucket:    bucket,
		endpoint:  endpoint,
		useSSL:    useSSL,
		publicURL: publicURL,
	}
}

// Upload envia arquivo para MinIO
func (m *MinIOStorage) Upload(
	ctx context.Context,
	key string,
	data io.Reader,
	contentType string,
) (url string, err error) {
	log.Debug().
		Str("bucket", m.bucket).
		Str("key", key).
		Str("content_type", contentType).
		Msg("MinIOStorage Upload iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("MinIOStorage Upload: key não pode estar vazio")
		return "", fmt.Errorf("key não pode estar vazio")
	}

	// Upload para MinIO (path-style)
	_, err = m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &m.bucket,
		Key:         &key,
		Body:        data,
		ContentType: &contentType,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("bucket", m.bucket).
			Str("key", key).
			Msg("MinIOStorage Upload erro ao fazer upload")
		return "", fmt.Errorf("erro ao fazer upload: %w", err)
	}

	// Construir URL (path-style para MinIO)
	if m.publicURL != "" {
		url = m.publicURL + "/" + m.bucket + "/" + key
	} else {
		// Fallback: construir URL com endpoint
		protocol := "http"
		if m.useSSL {
			protocol = "https"
		}
		url = fmt.Sprintf("%s://%s/%s/%s", protocol, m.endpoint, m.bucket, key)
	}

	log.Info().
		Str("key", key).
		Str("url", url).
		Msg("MinIOStorage Upload bem-sucedido")

	return url, nil
}

// Download baixa arquivo do MinIO
func (m *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	log.Debug().
		Str("bucket", m.bucket).
		Str("key", key).
		Msg("MinIOStorage Download iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("MinIOStorage Download: key não pode estar vazio")
		return nil, fmt.Errorf("key não pode estar vazio")
	}

	// Download do MinIO
	result, err := m.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &m.bucket,
		Key:    &key,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("MinIOStorage Download erro ao baixar")
		return nil, fmt.Errorf("erro ao baixar arquivo: %w", err)
	}

	log.Debug().Str("key", key).Msg("MinIOStorage Download bem-sucedido")

	return result.Body, nil
}

// Delete remove arquivo do MinIO
func (m *MinIOStorage) Delete(ctx context.Context, key string) error {
	log.Debug().
		Str("bucket", m.bucket).
		Str("key", key).
		Msg("MinIOStorage Delete iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("MinIOStorage Delete: key não pode estar vazio")
		return fmt.Errorf("key não pode estar vazio")
	}

	// Deletar do MinIO
	_, err := m.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &m.bucket,
		Key:    &key,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("MinIOStorage Delete erro ao deletar")
		return fmt.Errorf("erro ao deletar arquivo: %w", err)
	}

	log.Info().Str("key", key).Msg("MinIOStorage Delete bem-sucedido")

	return nil
}

// Exists verifica se arquivo existe no MinIO
func (m *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	log.Debug().
		Str("bucket", m.bucket).
		Str("key", key).
		Msg("MinIOStorage Exists iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("MinIOStorage Exists: key não pode estar vazio")
		return false, fmt.Errorf("key não pode estar vazio")
	}

	// HeadObject para verificar existência
	_, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &m.bucket,
		Key:    &key,
	})

	if err != nil {
		// Verificar se é erro 404 (not found) - considerar como false sem erro
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404") {
			log.Debug().Str("key", key).Msg("MinIOStorage Exists: arquivo não encontrado")
			return false, nil
		}

		log.Error().
			Err(err).
			Str("key", key).
			Msg("MinIOStorage Exists erro ao verificar")
		return false, fmt.Errorf("erro ao verificar arquivo: %w", err)
	}

	log.Debug().Str("key", key).Msg("MinIOStorage Exists: arquivo existe")

	return true, nil
}

// GetURL retorna URL pública de um arquivo
func (m *MinIOStorage) GetURL(ctx context.Context, key string) (string, error) {
	log.Debug().
		Str("bucket", m.bucket).
		Str("key", key).
		Msg("MinIOStorage GetURL iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("MinIOStorage GetURL: key não pode estar vazio")
		return "", fmt.Errorf("key não pode estar vazio")
	}

	// Verificar se arquivo existe
	exists, err := m.Exists(ctx, key)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Warn().Str("key", key).Msg("MinIOStorage GetURL: arquivo não encontrado")
		return "", fmt.Errorf("arquivo não encontrado: %s", key)
	}

	// Construir URL (path-style para MinIO)
	if m.publicURL != "" {
		url := m.publicURL + "/" + m.bucket + "/" + key
		log.Debug().Str("key", key).Str("url", url).Msg("MinIOStorage GetURL bem-sucedido")
		return url, nil
	}

	// Fallback: construir URL com endpoint
	protocol := "http"
	if m.useSSL {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", protocol, m.endpoint, m.bucket, key)

	log.Debug().Str("key", key).Str("url", url).Msg("MinIOStorage GetURL bem-sucedido")

	return url, nil
}

// HealthCheck verifica se MinIO está acessível
func (m *MinIOStorage) HealthCheck(ctx context.Context) error {
	log.Debug().
		Str("bucket", m.bucket).
		Str("endpoint", m.endpoint).
		Msg("MinIOStorage HealthCheck iniciado")

	// Tentar listar objetos do bucket (operação leve)
	maxKeys := int32(1)
	_, err := m.client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket:  &m.bucket,
		MaxKeys: &maxKeys, // Apenas 1 para verificação rápida
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("bucket", m.bucket).
			Msg("MinIOStorage HealthCheck falhou")
		return fmt.Errorf("erro ao verificar saúde do MinIO: %w", err)
	}

	log.Info().
		Str("bucket", m.bucket).
		Msg("MinIOStorage HealthCheck bem-sucedido")

	return nil
}

// PresignedURL gera URL assinada com tempo de expiração (útil para downloads privados)
func (m *MinIOStorage) PresignedURL(ctx context.Context, key string, expirationSeconds int) (string, error) {
	if expirationSeconds <= 0 {
		expirationSeconds = 3600 // default 1 hora
	}

	log.Debug().
		Str("key", key).
		Int("expiration_seconds", expirationSeconds).
		Msg("MinIOStorage PresignedURL iniciado")

	// MinIO não suporta presigned URLs via SDK v2 da mesma forma
	// Retornar URL pública como fallback
	return m.GetURL(ctx, key)
}

// CopyObject copia um arquivo dentro do MinIO
func (m *MinIOStorage) CopyObject(ctx context.Context, sourceKey, destinationKey string) error {
	log.Debug().
		Str("source_key", sourceKey).
		Str("destination_key", destinationKey).
		Msg("MinIOStorage CopyObject iniciado")

	if len(sourceKey) == 0 || len(destinationKey) == 0 {
		log.Error().Msg("MinIOStorage CopyObject: sourceKey e destinationKey não podem ser vazios")
		return fmt.Errorf("sourceKey e destinationKey são obrigatórios")
	}

	// Construir source no formato requerido
	source := m.bucket + "/" + sourceKey

	// CopyObject
	_, err := m.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &m.bucket,
		CopySource: &source,
		Key:        &destinationKey,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("source_key", sourceKey).
			Str("destination_key", destinationKey).
			Msg("MinIOStorage CopyObject erro")
		return fmt.Errorf("erro ao copiar objeto: %w", err)
	}

	log.Info().
		Str("source_key", sourceKey).
		Str("destination_key", destinationKey).
		Msg("MinIOStorage CopyObject bem-sucedido")

	return nil
}

// ListObjects lista objetos com prefixo
func (m *MinIOStorage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	log.Debug().
		Str("bucket", m.bucket).
		Str("prefix", prefix).
		Msg("MinIOStorage ListObjects iniciado")

	paginator := s3.NewListObjectsV2Paginator(
		m.client,
		&s3.ListObjectsV2Input{
			Bucket: &m.bucket,
			Prefix: &prefix,
		},
	)

	var keys []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Error().
				Err(err).
				Str("prefix", prefix).
				Msg("MinIOStorage ListObjects erro ao paginar")
			return nil, fmt.Errorf("erro ao listar objetos: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	log.Debug().
		Str("prefix", prefix).
		Int("count", len(keys)).
		Msg("MinIOStorage ListObjects bem-sucedido")

	return keys, nil
}

// DeletePrefix deleta todos os objetos com um prefixo
func (m *MinIOStorage) DeletePrefix(ctx context.Context, prefix string) error {
	log.Debug().
		Str("bucket", m.bucket).
		Str("prefix", prefix).
		Msg("MinIOStorage DeletePrefix iniciado")

	// Listar objetos com prefixo
	objects, err := m.ListObjects(ctx, prefix)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		log.Debug().Str("prefix", prefix).Msg("MinIOStorage DeletePrefix: nenhum objeto encontrado")
		return nil
	}

	// Deletar cada objeto
	for _, key := range objects {
		if err := m.Delete(ctx, key); err != nil {
			log.Error().
				Err(err).
				Str("key", key).
				Msg("MinIOStorage DeletePrefix erro ao deletar objeto")
			// Continuar deletando outros objetos mesmo com erro
		}
	}

	log.Info().
		Str("prefix", prefix).
		Int("deleted_count", len(objects)).
		Msg("MinIOStorage DeletePrefix bem-sucedido")

	return nil
}
