// internal/domain/entity/user.go
package entity

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// User representa a entidade de usuário no domínio
type User struct {
	ID                     uuid.UUID `gorm:"primaryKey" json:"id"`
	Name                   string    `gorm:"column:name" json:"name"`
	Email                  string    `gorm:"uniqueIndex;column:email" json:"email"`
	PasswordHash           string    `gorm:"column:password_hash" json:"-"`
	PlanID                 uuid.UUID `gorm:"index;column:plan_id" json:"planId"`
	ArticlesUsed           int       `gorm:"column:articles_used" json:"articlesUsed"`
	HasCompletedOnboarding bool      `gorm:"column:has_completed_onboarding" json:"hasCompletedOnboarding"`
	CreatedAt              time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt              time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName especifica o nome da tabela
func (User) TableName() string {
	return "users"
}

// Validate valida as regras de negócio da entidade User
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return errors.New("id é obrigatório")
	}

	if len(u.Name) < 2 || len(u.Name) > 100 {
		return errors.New("nome deve ter entre 2 e 100 caracteres")
	}

	if !isValidEmail(u.Email) {
		return errors.New("email inválido")
	}

	if u.PasswordHash == "" {
		return errors.New("password hash é obrigatório")
	}

	if u.PlanID == uuid.Nil {
		return errors.New("plan_id é obrigatório")
	}

	if u.ArticlesUsed < 0 {
		return errors.New("articles_used não pode ser negativo")
	}

	return nil
}

// CanGenerateArticles verifica se usuário pode gerar mais artigos
func (u *User) CanGenerateArticles(count int, maxArticles int) bool {
	return u.ArticlesUsed+count <= maxArticles
}

// IncrementArticlesUsed incrementa contador de artigos usados
func (u *User) IncrementArticlesUsed(count int) error {
	if count < 0 {
		return errors.New("count não pode ser negativo")
	}

	u.ArticlesUsed += count
	u.UpdatedAt = time.Now()
	return nil
}

// CompleteOnboarding marca onboarding como completo
func (u *User) CompleteOnboarding() {
	u.HasCompletedOnboarding = true
	u.UpdatedAt = time.Now()
}

// UpdatePlan atualiza o plano do usuário
func (u *User) UpdatePlan(planID uuid.UUID) error {
	if planID == uuid.Nil {
		return errors.New("plan_id inválido")
	}

	u.PlanID = planID
	u.UpdatedAt = time.Now()
	return nil
}

// isValidEmail valida formato de email (RFC 5322 simplificado)
func isValidEmail(email string) bool {
	// Padrão simplificado de email
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// ============================================
// REFRESH TOKEN ENTITY
// ============================================

// RefreshToken representa um token de refresh armazenado
type RefreshToken struct {
	ID         uuid.UUID  `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"index;column:user_id" json:"userId"`
	TokenHash  string     `gorm:"uniqueIndex;column:token_hash" json:"-"`
	ExpiresAt  time.Time  `gorm:"index;column:expires_at" json:"expiresAt"`
	LastUsedAt *time.Time `gorm:"column:last_used_at" json:"lastUsedAt"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"createdAt"`
}

// TableName especifica o nome da tabela
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired verifica se o token expirou
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// UpdateLastUsed atualiza timestamp de último uso
func (rt *RefreshToken) UpdateLastUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}
