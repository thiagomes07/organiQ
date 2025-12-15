// internal/domain/entity/business_profile.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// BusinessProfile representa o perfil de negócio do usuário
type BusinessProfile struct {
	ID                 uuid.UUID      `gorm:"primaryKey" json:"id"`
	UserID             uuid.UUID      `gorm:"uniqueIndex;column:user_id" json:"userId"`
	Description        string         `gorm:"type:text;column:description" json:"description"`
	PrimaryObjective   Objective      `gorm:"column:primary_objective" json:"primaryObjective"`
	SecondaryObjective *Objective     `gorm:"column:secondary_objective" json:"secondaryObjective"`
	Location           Location       `gorm:"type:jsonb;column:location" json:"location"`
	SiteURL            *string        `gorm:"column:site_url" json:"siteUrl"`
	HasBlog            bool           `gorm:"column:has_blog" json:"hasBlog"`
	BlogURLs           BlogURLs       `gorm:"type:jsonb;column:blog_urls" json:"blogUrls"`
	BrandFileURL       *string        `gorm:"column:brand_file_url" json:"brandFileUrl"`
	CreatedAt          time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          time.Time      `gorm:"column:updated_at" json:"updatedAt"`
}

// Objective enum para objetivos de negócio
type Objective string

const (
	ObjectiveLeads    Objective = "leads"
	ObjectiveSales    Objective = "sales"
	ObjectiveBranding Objective = "branding"
)

// IsValid verifica se o objetivo é válido
func (o Objective) IsValid() bool {
	return o == ObjectiveLeads || o == ObjectiveSales || o == ObjectiveBranding
}

// Location estrutura para localização geográfica
type Location struct {
	Country           string `json:"country"`
	State             string `json:"state"`
	City              string `json:"city"`
	HasMultipleUnits  bool   `json:"hasMultipleUnits"`
	Units             []Unit `json:"units"`
}

// Unit representa uma unidade de negócio
type Unit struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Country string    `json:"country"`
	State   string    `json:"state"`
	City    string    `json:"city"`
}

// BlogURLs representa lista de URLs de blogs
type BlogURLs []string

// Scan implementa sql.Scanner para Location
func (l *Location) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed for Location")
	}
	return json.Unmarshal(bytes, &l)
}

// Value implementa driver.Valuer para Location
func (l Location) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// Scan implementa sql.Scanner para BlogURLs
func (b *BlogURLs) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed for BlogURLs")
	}
	return json.Unmarshal(bytes, &b)
}

// Value implementa driver.Valuer para BlogURLs
func (b BlogURLs) Value() (driver.Value, error) {
	if len(b) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(b)
}

// TableName especifica o nome da tabela
func (BusinessProfile) TableName() string {
	return "business_profiles"
}

// Validate valida as regras de negócio
func (bp *BusinessProfile) Validate() error {
	if bp.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if bp.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if len(bp.Description) == 0 || len(bp.Description) > 500 {
		return errors.New("description deve ter entre 1 e 500 caracteres")
	}

	if !bp.PrimaryObjective.IsValid() {
		return errors.New("primaryObjective inválido")
	}

	if bp.SecondaryObjective != nil && !bp.SecondaryObjective.IsValid() {
		return errors.New("secondaryObjective inválido")
	}

	if err := bp.Location.Validate(); err != nil {
		return err
	}

	if bp.SiteURL != nil && len(*bp.SiteURL) == 0 {
		return errors.New("siteUrl deve ser não-vazio se fornecido")
	}

	if bp.HasBlog && len(bp.BlogURLs) == 0 {
		return errors.New("blogUrls é obrigatório quando hasBlog é true")
	}

	if bp.BrandFileURL != nil && len(*bp.BrandFileURL) == 0 {
		return errors.New("brandFileUrl deve ser não-vazio se fornecido")
	}

	return nil
}

// Validate valida a estrutura Location
func (l Location) Validate() error {
	if len(l.Country) == 0 {
		return errors.New("location.country é obrigatório")
	}

	if len(l.State) == 0 {
		return errors.New("location.state é obrigatório")
	}

	if len(l.City) == 0 {
		return errors.New("location.city é obrigatório")
	}

	if l.HasMultipleUnits && len(l.Units) == 0 {
		return errors.New("units é obrigatório quando hasMultipleUnits é true")
	}

	for _, unit := range l.Units {
		if err := unit.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate valida a estrutura Unit
func (u Unit) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("unit.id é obrigatório")
	}

	if len(u.Name) == 0 {
		return errors.New("unit.name é obrigatório")
	}

	if len(u.Country) == 0 {
		return errors.New("unit.country é obrigatório")
	}

	if len(u.State) == 0 {
		return errors.New("unit.state é obrigatório")
	}

	if len(u.City) == 0 {
		return errors.New("unit.city é obrigatório")
	}

	return nil
}

// AddBlogURL adiciona uma URL de blog
func (bp *BusinessProfile) AddBlogURL(url string) error {
	if len(url) == 0 {
		return errors.New("url não pode estar vazio")
	}

	for _, existing := range bp.BlogURLs {
		if existing == url {
			return errors.New("url já existe")
		}
	}

	bp.BlogURLs = append(bp.BlogURLs, url)
	bp.UpdatedAt = time.Now()
	return nil
}

// RemoveBlogURL remove uma URL de blog
func (bp *BusinessProfile) RemoveBlogURL(url string) {
	filtered := make(BlogURLs, 0, len(bp.BlogURLs))
	for _, existing := range bp.BlogURLs {
		if existing != url {
			filtered = append(filtered, existing)
		}
	}
	bp.BlogURLs = filtered
	bp.UpdatedAt = time.Now()
}

// SetPrimaryObjective altera o objetivo primário
func (bp *BusinessProfile) SetPrimaryObjective(objective Objective) error {
	if !objective.IsValid() {
		return errors.New("objetivo inválido")
	}
	bp.PrimaryObjective = objective
	bp.UpdatedAt = time.Now()
	return nil
}

// SetSecondaryObjective altera o objetivo secundário
func (bp *BusinessProfile) SetSecondaryObjective(objective *Objective) error {
	if objective != nil && !objective.IsValid() {
		return errors.New("objetivo inválido")
	}
	bp.SecondaryObjective = objective
	bp.UpdatedAt = time.Now()
	return nil
}
