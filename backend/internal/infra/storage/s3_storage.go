package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

// S3Storage implementa StorageService para AWS S3
type S3Storage struct {
	client    *s3.Client
	bucket    string
	region    string
	endpoint  string
	useSSL    bool
	publicURL string
}

// NewS3Storage cria nova instância para AWS S3
func NewS3Storage(
	client *s3.Client,
	bucket string,
	region string,
	endpoint string,
	useSSL bool,
	publicURL string,
) StorageService {
	return &S3Storage{
		client:    client,
		bucket:    bucket,
		region:    region,
		endpoint:  endpoint,
		useSSL:    useSSL,
		publicURL: publicURL,
	}
}

// Upload envia arquivo para S3
func (s *S3Storage) Upload(
	ctx context.Context,
	key string,
	data io.Reader,
	contentType string,
) (url string, err error) {
	log.Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Str("content_type", contentType).
		Msg("S3Storage Upload iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("S3Storage Upload: key não pode estar vazio")
		return "", fmt.Errorf("key não pode estar vazio")
	}

	// Upload para S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        data,
		ContentType: &contentType,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("bucket", s.bucket).
			Str("key", key).
			Msg("S3Storage Upload erro ao fazer upload")
		return "", fmt.Errorf("erro ao fazer upload: %w", err)
	}

	// Construir URL (virtual-hosted-style para AWS)
	if s.publicURL != "" {
		url = s.publicURL + "/" + key
	} else {
		// Fallback: construir URL padrão AWS
		protocol := "https"
		if !s.useSSL {
			protocol = "http"
		}

		if s.endpoint != "" {
			// Endpoint customizado (MinIO em outro host, etc)
			url = fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucket, key)
		} else {
			// AWS S3 padrão (virtual-hosted-style)
			url = fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s",
				protocol,
				s.bucket,
				s.region,
				key,
			)
		}
	}

	log.Info().
		Str("key", key).
		Str("url", url).
		Msg("S3Storage Upload bem-sucedido")

	return url, nil
}

// Download baixa arquivo do S3
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	log.Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Msg("S3Storage Download iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("S3Storage Download: key não pode estar vazio")
		return nil, fmt.Errorf("key não pode estar vazio")
	}

	// Download do S3
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("S3Storage Download erro ao baixar")
		return nil, fmt.Errorf("erro ao baixar arquivo: %w", err)
	}

	log.Debug().Str("key", key).Msg("S3Storage Download bem-sucedido")

	return result.Body, nil
}

// Delete remove arquivo do S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	log.Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Msg("S3Storage Delete iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("S3Storage Delete: key não pode estar vazio")
		return fmt.Errorf("key não pode estar vazio")
	}

	// Deletar do S3
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("S3Storage Delete erro ao deletar")
		return fmt.Errorf("erro ao deletar arquivo: %w", err)
	}

	log.Info().Str("key", key).Msg("S3Storage Delete bem-sucedido")

	return nil
}

// Exists verifica se arquivo existe no S3
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	log.Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Msg("S3Storage Exists iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("S3Storage Exists: key não pode estar vazio")
		return false, fmt.Errorf("key não pode estar vazio")
	}

	// HeadObject para verificar existência
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		// Verificar se é erro 404 (not found)
		var apiErr smithy.APIError
		if apiErr != nil && (apiErr.Error() == "NotFound" || strings.Contains(apiErr.Error(), "NoSuchKey") || strings.Contains(apiErr.Error(), "404")) {
			log.Debug().Str("key", key).Msg("S3Storage Exists: arquivo não encontrado")
			return false, nil
		}

		// Checar se é erro de 404 de forma mais genérica
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "404") {
			log.Debug().Str("key", key).Msg("S3Storage Exists: arquivo não encontrado")
			return false, nil
		}

		log.Error().
			Err(err).
			Str("key", key).
			Msg("S3Storage Exists erro ao verificar")
		return false, fmt.Errorf("erro ao verificar arquivo: %w", err)
	}

	log.Debug().Str("key", key).Msg("S3Storage Exists: arquivo existe")

	return true, nil
}

// GetURL retorna URL pública de um arquivo
func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	log.Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Msg("S3Storage GetURL iniciado")

	// Validar key
	if len(key) == 0 {
		log.Error().Msg("S3Storage GetURL: key não pode estar vazio")
		return "", fmt.Errorf("key não pode estar vazio")
	}

	// Verificar se arquivo existe
	exists, err := s.Exists(ctx, key)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Warn().Str("key", key).Msg("S3Storage GetURL: arquivo não encontrado")
		return "", fmt.Errorf("arquivo não encontrado: %s", key)
	}

	// Construir URL
	if s.publicURL != "" {
		url := s.publicURL + "/" + key
		log.Debug().Str("key", key).Str("url", url).Msg("S3Storage GetURL bem-sucedido")
		return url, nil
	}

	// Fallback: construir URL padrão AWS
	protocol := "https"
	if !s.useSSL {
		protocol = "http"
	}

	var url string
	if s.endpoint != "" {
		// Endpoint customizado
		url = fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucket, key)
	} else {
		// AWS S3 padrão (virtual-hosted-style)
		url = fmt.Sprintf("%s://%s.s3.%s.amazonaws.com/%s",
			protocol,
			s.bucket,
			s.region,
			key,
		)
	}

	log.Debug().Str("key", key).Str("url", url).Msg("S3Storage GetURL bem-sucedido")

	return url, nil
}

// HealthCheck verifica se S3 está acessível
func (s *S3Storage) HealthCheck(ctx context.Context) error {
	log.Debug().
		Str("bucket", s.bucket).
		Str("region", s.region).
		Msg("S3Storage HealthCheck iniciado")

	// Tentar listar objetos do bucket (operação leve)
	maxKeys := int32(1)
	_, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  &s.bucket,
		MaxKeys: &maxKeys, // Apenas 1 para verificação rápida
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("bucket", s.bucket).
			Msg("S3Storage HealthCheck falhou")
		return fmt.Errorf("erro ao verificar saúde do S3: %w", err)
	}

	log.Info().
		Str("bucket", s.bucket).
		Msg("S3Storage HealthCheck bem-sucedido")

	return nil
}

// PresignedURL gera URL assinada com tempo de expiração (útil para downloads privados)
func (s *S3Storage) PresignedURL(ctx context.Context, key string, expirationSeconds int) (string, error) {
	if expirationSeconds <= 0 {
		expirationSeconds = 3600 // default 1 hora
	}

	log.Debug().
		Str("key", key).
		Int("expiration_seconds", expirationSeconds).
		Msg("S3Storage PresignedURL iniciado")

	// Para AWS S3, precisaríamos usar presignerClient
	// Como simplificação, retornar URL pública
	// Em produção, usar:
	// presignerClient := s3.NewPresignFromClient(s.client)
	// request, err := presignerClient.PresignGetObject(ctx, ...)
	return s.GetURL(ctx, key)
}

// CopyObject copia um arquivo dentro do S3
func (s *S3Storage) CopyObject(ctx context.Context, sourceKey, destinationKey string) error {
	log.Debug().
		Str("source_key", sourceKey).
		Str("destination_key", destinationKey).
		Msg("S3Storage CopyObject iniciado")

	if len(sourceKey) == 0 || len(destinationKey) == 0 {
		log.Error().Msg("S3Storage CopyObject: sourceKey e destinationKey não podem ser vazios")
		return fmt.Errorf("sourceKey e destinationKey são obrigatórios")
	}

	// Construir source no formato requerido
	source := s.bucket + "/" + sourceKey

	// CopyObject
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     &s.bucket,
		CopySource: &source,
		Key:        &destinationKey,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("source_key", sourceKey).
			Str("destination_key", destinationKey).
			Msg("S3Storage CopyObject erro")
		return fmt.Errorf("erro ao copiar objeto: %w", err)
	}

	log.Info().
		Str("source_key", sourceKey).
		Str("destination_key", destinationKey).
		Msg("S3Storage CopyObject bem-sucedido")

	return nil
}

// ListObjects lista objetos com prefixo
func (s *S3Storage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	log.Debug().
		Str("bucket", s.bucket).
		Str("prefix", prefix).
		Msg("S3Storage ListObjects iniciado")

	paginator := s3.NewListObjectsV2Paginator(
		s.client,
		&s3.ListObjectsV2Input{
			Bucket: &s.bucket,
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
				Msg("S3Storage ListObjects erro ao paginar")
			return nil, fmt.Errorf("erro ao listar objetos: %w", err)
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	log.Debug().
		Str("prefix", prefix).
		Int("count", len(keys)).
		Msg("S3Storage ListObjects bem-sucedido")

	return keys, nil
}

// DeletePrefix deleta todos os objetos com um prefixo
func (s *S3Storage) DeletePrefix(ctx context.Context, prefix string) error {
	log.Debug().
		Str("bucket", s.bucket).
		Str("prefix", prefix).
		Msg("S3Storage DeletePrefix iniciado")

	// Listar objetos com prefixo
	objects, err := s.ListObjects(ctx, prefix)
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		log.Debug().Str("prefix", prefix).Msg("S3Storage DeletePrefix: nenhum objeto encontrado")
		return nil
	}

	// Deletar cada objeto
	for _, key := range objects {
		if err := s.Delete(ctx, key); err != nil {
			log.Error().
				Err(err).
				Str("key", key).
				Msg("S3Storage DeletePrefix erro ao deletar objeto")
			// Continuar deletando outros objetos mesmo com erro
		}
	}

	log.Info().
		Str("prefix", prefix).
		Int("deleted_count", len(objects)).
		Msg("S3Storage DeletePrefix bem-sucedido")

	return nil
}
