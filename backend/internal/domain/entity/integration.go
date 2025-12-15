// internal/domain/entity/integration.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Integration representa uma integração de terceiros
type Integration struct {
	ID        uuid.UUID          `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID          `gorm:"index;column:user_id" json:"userId"`
	Type      IntegrationType    `gorm:"index;column:type" json:"type"`
	Config    IntegrationConfig  `gorm:"type:jsonb;column:config" json:"config"`
	Enabled   bool               `gorm:"column:enabled" json:"enabled"`
	CreatedAt time.Time          `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time          `gorm:"column:updated_at" json:"updatedAt"`
}

// IntegrationType enum para tipos de integração
type IntegrationType string

const (
	IntegrationTypeWordPress     IntegrationType = "wordpress"
	IntegrationTypeSearchConsole IntegrationType = "search_console"
	IntegrationTypeAnalytics     IntegrationType = "analytics"
)

// IsValid verifica se o tipo é válido
func (it IntegrationType) IsValid() bool {
	return it == IntegrationTypeWordPress ||
		it == IntegrationTypeSearchConsole ||
		it == IntegrationTypeAnalytics
}

// IntegrationConfig representa a configuração genérica (será específica por tipo)
type IntegrationConfig map[string]interface{}

// Scan implementa sql.Scanner para IntegrationConfig
func (ic *IntegrationConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion failed for IntegrationConfig")
	}
	return json.Unmarshal(bytes, &ic)
}

// Value implementa driver.Valuer para IntegrationConfig
func (ic IntegrationConfig) Value() (driver.Value, error) {
	return json.Marshal(ic)
}

// TableName especifica o nome da tabela
func (Integration) TableName() string {
	return "integrations"
}

// Validate valida as regras de negócio
func (i *Integration) Validate() error {
	if i.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if i.UserID == uuid.Nil {
		return errors.New("user_id é obrigatório")
	}

	if !i.Type.IsValid() {
		return errors.New("tipo de integração inválido")
	}

	if len(i.Config) == 0 {
		return errors.New("config não pode estar vazio")
	}

	switch i.Type {
	case IntegrationTypeWordPress:
		return i.validateWordPressConfig()
	case IntegrationTypeSearchConsole:
		return i.validateSearchConsoleConfig()
	case IntegrationTypeAnalytics:
		return i.validateAnalyticsConfig()
	}

	return nil
}

// validateWordPressConfig valida configuração de WordPress
func (i *Integration) validateWordPressConfig() error {
	siteURL, ok := i.Config["siteUrl"].(string)
	if !ok || len(siteURL) == 0 {
		return errors.New("wordpress config: siteUrl é obrigatório")
	}

	username, ok := i.Config["username"].(string)
	if !ok || len(username) == 0 {
		return errors.New("wordpress config: username é obrigatório")
	}

	appPassword, ok := i.Config["appPassword"].(string)
	if !ok || len(appPassword) == 0 {
		return errors.New("wordpress config: appPassword é obrigatório")
	}

	return nil
}

// validateSearchConsoleConfig valida configuração de Search Console
func (i *Integration) validateSearchConsoleConfig() error {
	propertyURL, ok := i.Config["propertyUrl"].(string)
	if !ok || len(propertyURL) == 0 {
		return errors.New("search_console config: propertyUrl é obrigatório")
	}

	return nil
}

// validateAnalyticsConfig valida configuração de Analytics
func (i *Integration) validateAnalyticsConfig() error {
	measurementID, ok := i.Config["measurementId"].(string)
	if !ok || len(measurementID) == 0 {
		return errors.New("analytics config: measurementId é obrigatório")
	}

	return nil
}

// GetWordPressConfig extrai configuração de WordPress com segurança
func (i *Integration) GetWordPressConfig() (*WordPressConfig, error) {
	if i.Type != IntegrationTypeWordPress {
		return nil, errors.New("integração não é tipo wordpress")
	}

	siteURL, _ := i.Config["siteUrl"].(string)
	username, _ := i.Config["username"].(string)
	appPassword, _ := i.Config["appPassword"].(string)

	return &WordPressConfig{
		SiteURL:    siteURL,
		Username:   username,
		AppPassword: appPassword,
	}, nil
}

// GetSearchConsoleConfig extrai configuração de Search Console
func (i *Integration) GetSearchConsoleConfig() (*SearchConsoleConfig, error) {
	if i.Type != IntegrationTypeSearchConsole {
		return nil, errors.New("integração não é tipo search_console")
	}

	propertyURL, _ := i.Config["propertyUrl"].(string)

	return &SearchConsoleConfig{
		PropertyURL: propertyURL,
	}, nil
}

// GetAnalyticsConfig extrai configuração de Analytics
func (i *Integration) GetAnalyticsConfig() (*AnalyticsConfig, error) {
	if i.Type != IntegrationTypeAnalytics {
		return nil, errors.New("integração não é tipo analytics")
	}

	measurementID, _ := i.Config["measurementId"].(string)

	return &AnalyticsConfig{
		MeasurementID: measurementID,
	}, nil
}

// SetWordPressConfig define configuração de WordPress
func (i *Integration) SetWordPressConfig(config *WordPressConfig) error {
	if config == nil {
		return errors.New("config não pode ser nil")
	}

	if len(config.SiteURL) == 0 {
		return errors.New("siteUrl é obrigatório")
	}

	if len(config.Username) == 0 {
		return errors.New("username é obrigatório")
	}

	if len(config.AppPassword) == 0 {
		return errors.New("appPassword é obrigatório")
	}

	i.Type = IntegrationTypeWordPress
	i.Config = IntegrationConfig{
		"siteUrl":    config.SiteURL,
		"username":   config.Username,
		"appPassword": config.AppPassword,
	}
	i.UpdatedAt = time.Now()
	return nil
}

// SetSearchConsoleConfig define configuração de Search Console
func (i *Integration) SetSearchConsoleConfig(config *SearchConsoleConfig) error {
	if config == nil {
		return errors.New("config não pode ser nil")
	}

	if len(config.PropertyURL) == 0 {
		return errors.New("propertyUrl é obrigatório")
	}

	i.Type = IntegrationTypeSearchConsole
	i.Config = IntegrationConfig{
		"propertyUrl": config.PropertyURL,
	}
	i.UpdatedAt = time.Now()
	return nil
}

// SetAnalyticsConfig define configuração de Analytics
func (i *Integration) SetAnalyticsConfig(config *AnalyticsConfig) error {
	if config == nil {
		return errors.New("config não pode ser nil")
	}

	if len(config.MeasurementID) == 0 {
		return errors.New("measurementId é obrigatório")
	}

	i.Type = IntegrationTypeAnalytics
	i.Config = IntegrationConfig{
		"measurementId": config.MeasurementID,
	}
	i.UpdatedAt = time.Now()
	return nil
}

// Enable ativa a integração
func (i *Integration) Enable() {
	i.Enabled = true
	i.UpdatedAt = time.Now()
}

// Disable desativa a integração
func (i *Integration) Disable() {
	i.Enabled = false
	i.UpdatedAt = time.Now()
}

// ============================================
// CONFIG STRUCTS
// ============================================

// WordPressConfig configuração específica do WordPress
type WordPressConfig struct {
	SiteURL    string `json:"siteUrl"`
	Username   string `json:"username"`
	AppPassword string `json:"appPassword"`
}

// SearchConsoleConfig configuração específica do Search Console
type SearchConsoleConfig struct {
	PropertyURL string `json:"propertyUrl"`
}

// AnalyticsConfig configuração específica do Analytics
type AnalyticsConfig struct {
	MeasurementID string `json:"measurementId"`
}
