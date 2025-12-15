package util

import (
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validatorOnce   sync.Once
	structValidator *validator.Validate
)

// EnsureValidator lazily initializes the struct validator.
func ensureValidator() {
	validatorOnce.Do(func() {
		structValidator = validator.New(validator.WithRequiredStructEnabled())
	})
}

// IsValidEmail valida o formato básico de email.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}

// ValidateStruct executa validação estrutural padronizada usando validator/v10.
func ValidateStruct(s interface{}) error {
	if s == nil {
		return nil
	}

	ensureValidator()
	return structValidator.Struct(s)
}
