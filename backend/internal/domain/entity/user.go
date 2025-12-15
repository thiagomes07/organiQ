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
	ID                     uuid.UUID
	Name                   string
	Email                  string
	PasswordHash           string
	PlanID                 uuid.UUID
	ArticlesUsed           int
	HasCompletedOnboarding bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
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
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string // Hash SHA-256 do token
	ExpiresAt  time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
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
