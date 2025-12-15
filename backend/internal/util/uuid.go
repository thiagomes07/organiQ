package util

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// MustParseUUID faz parse de string para UUID, retorna uuid.Nil se inválido
// Útil para conversões sem muita cerimônia
func MustParseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}

	parsed, err := uuid.Parse(s)
	if err != nil {
		log.Debug().
			Str("input", s).
			Err(err).
			Msg("MustParseUUID: erro ao fazer parse, retornando uuid.Nil")
		return uuid.Nil
	}

	return parsed
}

// ParseUUID faz parse de string para UUID, retorna erro se inválido
func ParseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, nil
	}

	return uuid.Parse(s)
}

// UUIDToString converte UUID para string
func UUIDToString(id uuid.UUID) string {
	return id.String()
}

// IsValidUUID valida se string é um UUID válido
func IsValidUUID(s string) bool {
	if s == "" {
		return false
	}

	_, err := uuid.Parse(s)
	return err == nil
}
