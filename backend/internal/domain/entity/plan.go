// internal/domain/entity/plan.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Plan representa um plano de assinatura
type Plan struct {
	ID                          uuid.UUID `gorm:"primaryKey" json:"id"`
	Name                        string    `gorm:"uniqueIndex;column:name" json:"name"`
	MaxArticles                 int       `gorm:"column:max_articles" json:"maxArticles"`
	MaxIdeaRegenerationsPerHour int       `gorm:"column:max_idea_regenerations_per_hour" json:"maxIdeaRegenerationsPerHour"`
	Price                       float64   `gorm:"column:price" json:"price"`
	Features                    Features  `gorm:"type:jsonb;column:features" json:"features"`
	Active                      bool      `gorm:"column:active" json:"active"`
	CreatedAt                   time.Time `gorm:"column:created_at" json:"createdAt"`
}

// Features representa a lista de features do plano
type Features []string

// Scan implementa a interface sql.Scanner para GORM
func (f *Features) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed for Features")
	}
	return json.Unmarshal(bytes, &f)
}

// Value implementa a interface driver.Valuer para GORM
func (f Features) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// TableName especifica o nome da tabela
func (Plan) TableName() string {
	return "plans"
}

// Validate valida as regras de negócio do plano
func (p *Plan) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if len(p.Name) == 0 || len(p.Name) > 50 {
		return errors.New("nome do plano deve ter entre 1 e 50 caracteres")
	}

	if p.MaxArticles < 0 {
		return errors.New("maxArticles não pode ser negativo")
	}

	if p.Price < 0 {
		return errors.New("price não pode ser negativo")
	}

	if len(p.Features) == 0 {
		return errors.New("features não pode estar vazio")
	}

	return nil
}

// IsFreePlan verifica se o plano é o plano Free
func (p *Plan) IsFreePlan() bool {
	return p.Name == "Free"
}

// CanPublishArticles verifica se o plano permite publicar artigos
func (p *Plan) CanPublishArticles() bool {
	return p.MaxArticles > 0
}

// HasFeature verifica se o plano possui uma feature específica
func (p *Plan) HasFeature(feature string) bool {
	for _, f := range p.Features {
		if f == feature {
			return true
		}
	}
	return false
}
