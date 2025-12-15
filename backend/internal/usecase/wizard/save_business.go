// internal/usecase/wizard/save_business.go
package wizard

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"organiq/internal/domain/entity"
	"organiq/internal/domain/repository"
	"organiq/internal/infra/storage"
)

// SaveBusinessInput dados de entrada para salvar negócio
type SaveBusinessInput struct {
	UserID             string // UUID como string do context
	Description        string
	PrimaryObjective   string // "leads", "sales", "branding"
	SecondaryObjective *string
	Location           *entity.Location
	SiteURL            *string
	HasBlog            bool
	BlogURLs           []string
	BrandFile          io.Reader // File reader
	BrandFileName      string    // Original filename
	BrandFileSize      int64     // Tamanho do arquivo
}

// SaveBusinessOutput dados de saída
type SaveBusinessOutput struct {
	Success      bool
	ProfileID    string
	BrandFileURL *string
}

// SaveBusinessUseCase implementa o caso de uso
type SaveBusinessUseCase struct {
	businessRepo   repository.BusinessRepository
	storageService storage.StorageService
}

// NewSaveBusinessUseCase cria nova instância
func NewSaveBusinessUseCase(
	businessRepo repository.BusinessRepository,
	storageService storage.StorageService,
) *SaveBusinessUseCase {
	return &SaveBusinessUseCase{
		businessRepo:   businessRepo,
		storageService: storageService,
	}
}

// Execute executa o caso de uso
func (uc *SaveBusinessUseCase) Execute(ctx context.Context, input SaveBusinessInput) (*SaveBusinessOutput, error) {
	log.Debug().Str("user_id", input.UserID).Msg("SaveBusinessUseCase Execute iniciado")

	// 1. Parse user_id
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		log.Error().Err(err).Msg("SaveBusinessUseCase: user_id inválido")
		return nil, errors.New("invalid_user_id")
	}

	// 2. Validar entrada
	if len(input.Description) == 0 || len(input.Description) > 500 {
		log.Warn().Msg("SaveBusinessUseCase: description inválida")
		return nil, errors.New("description deve ter entre 1 e 500 caracteres")
	}

	primaryObjective := entity.Objective(input.PrimaryObjective)
	if !primaryObjective.IsValid() {
		log.Warn().Str("objective", input.PrimaryObjective).Msg("SaveBusinessUseCase: primaryObjective inválido")
		return nil, errors.New("primaryObjective inválido: deve ser 'leads', 'sales' ou 'branding'")
	}

	var secondaryObjective *entity.Objective
	if input.SecondaryObjective != nil {
		obj := entity.Objective(*input.SecondaryObjective)
		if !obj.IsValid() {
			log.Warn().Str("objective", *input.SecondaryObjective).Msg("SaveBusinessUseCase: secondaryObjective inválido")
			return nil, errors.New("secondaryObjective inválido")
		}
		secondaryObjective = &obj
	}

	// 3. Validar localização
	if input.Location == nil {
		log.Warn().Msg("SaveBusinessUseCase: location obrigatória")
		return nil, errors.New("location é obrigatório")
	}

	if err := input.Location.Validate(); err != nil {
		log.Warn().Err(err).Msg("SaveBusinessUseCase: location inválida")
		return nil, err
	}

	// 4. Validar URLs de blog se fornecidas
	if input.HasBlog && len(input.BlogURLs) == 0 {
		log.Warn().Msg("SaveBusinessUseCase: blogUrls obrigatório quando hasBlog=true")
		return nil, errors.New("blogUrls é obrigatório quando hasBlog é true")
	}

	// 5. Upload do arquivo de brand se fornecido
	var brandFileURL *string
	if input.BrandFile != nil && input.BrandFileSize > 0 {
		log.Debug().
			Str("user_id", input.UserID).
			Str("filename", input.BrandFileName).
			Int64("size", input.BrandFileSize).
			Msg("SaveBusinessUseCase: iniciando upload de brand file")

		// Validar tamanho (máximo 5MB)
		if input.BrandFileSize > 5*1024*1024 {
			log.Warn().Int64("size", input.BrandFileSize).Msg("SaveBusinessUseCase: arquivo muito grande")
			return nil, errors.New("arquivo não pode exceder 5MB")
		}

		// Validar tipo de arquivo
		if !isValidBrandFileType(input.BrandFileName) {
			log.Warn().Str("filename", input.BrandFileName).Msg("SaveBusinessUseCase: tipo de arquivo inválido")
			return nil, errors.New("arquivo deve ser PDF, JPG ou PNG")
		}

		// Gerar chave única para o arquivo
		fileKey := generateBrandFileKey(userID, input.BrandFileName)

		// Upload para storage
		url, err := uc.storageService.Upload(
			ctx,
			fileKey,
			input.BrandFile,
			getMimeType(input.BrandFileName),
		)
		if err != nil {
			log.Error().Err(err).Msg("SaveBusinessUseCase: erro ao fazer upload de brand file")
			return nil, errors.New("erro ao fazer upload do arquivo de marca")
		}

		brandFileURL = &url
		log.Info().Str("user_id", input.UserID).Str("url", url).Msg("SaveBusinessUseCase: upload bem-sucedido")
	}

	// 6. Buscar profile existente (pode ter sido criado em outro wizard step)
	existingProfile, err := uc.businessRepo.FindProfileByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("SaveBusinessUseCase: erro ao buscar profile existente")
		return nil, errors.New("erro ao buscar profile existente")
	}

	var profileID string

	if existingProfile != nil {
		// Atualizar profile existente
		log.Debug().Str("user_id", input.UserID).Msg("SaveBusinessUseCase: atualizando profile existente")

		existingProfile.Description = input.Description
		existingProfile.PrimaryObjective = primaryObjective
		existingProfile.SecondaryObjective = secondaryObjective
		existingProfile.Location = *input.Location
		existingProfile.SiteURL = input.SiteURL
		existingProfile.HasBlog = input.HasBlog
		existingProfile.BlogURLs = input.BlogURLs
		if brandFileURL != nil {
			existingProfile.BrandFileURL = brandFileURL
		}
		existingProfile.UpdatedAt = time.Now()

		if err := existingProfile.Validate(); err != nil {
			log.Error().Err(err).Msg("SaveBusinessUseCase: profile inválido após atualização")
			return nil, err
		}

		if err := uc.businessRepo.UpdateProfile(ctx, existingProfile); err != nil {
			log.Error().Err(err).Msg("SaveBusinessUseCase: erro ao atualizar profile")
			return nil, errors.New("erro ao salvar perfil de negócio")
		}

		profileID = existingProfile.ID.String()
	} else {
		// Criar novo profile
		log.Debug().Str("user_id", input.UserID).Msg("SaveBusinessUseCase: criando novo profile")

		profile := &entity.BusinessProfile{
			ID:                 uuid.New(),
			UserID:             userID,
			Description:        input.Description,
			PrimaryObjective:   primaryObjective,
			SecondaryObjective: secondaryObjective,
			Location:           *input.Location,
			SiteURL:            input.SiteURL,
			HasBlog:            input.HasBlog,
			BlogURLs:           input.BlogURLs,
			BrandFileURL:       brandFileURL,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := profile.Validate(); err != nil {
			log.Error().Err(err).Msg("SaveBusinessUseCase: profile inválido")
			return nil, err
		}

		if err := uc.businessRepo.CreateProfile(ctx, profile); err != nil {
			log.Error().Err(err).Msg("SaveBusinessUseCase: erro ao criar profile")
			return nil, errors.New("erro ao salvar perfil de negócio")
		}

		profileID = profile.ID.String()
	}

	log.Info().Str("user_id", input.UserID).Str("profile_id", profileID).Msg("SaveBusinessUseCase bem-sucedido")

	return &SaveBusinessOutput{
		Success:      true,
		ProfileID:    profileID,
		BrandFileURL: brandFileURL,
	}, nil
}

// ============================================
// HELPERS
// ============================================

func generateBrandFileKey(userID uuid.UUID, originalFilename string) string {
	// Formato: brand-files/{userID}/{uniqueID}_{originalFilename}
	return "brand-files/" + userID.String() + "/" + uuid.New().String() + "_" + originalFilename
}

func isValidBrandFileType(filename string) bool {
	// Validar extensão
	validExtensions := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	// Extrair extensão
	var ext string
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = filename[i:]
			break
		}
	}

	if ext == "" {
		return false
	}

	// Comparar (case-insensitive)
	for valid := range validExtensions {
		if toLower(ext) == valid {
			return true
		}
	}

	return false
}

func getMimeType(filename string) string {
	// Extrair extensão
	var ext string
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			ext = toLower(filename[i:])
			break
		}
	}

	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
