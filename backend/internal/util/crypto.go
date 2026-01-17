// internal/util/crypto.go
package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// CryptoService encapsula operações de criptografia
type CryptoService struct {
	passwordPepper string
	jwtSecret      string
}

// NewCryptoService cria nova instância do serviço de crypto
func NewCryptoService(passwordPepper, jwtSecret string) *CryptoService {
	return &CryptoService{
		passwordPepper: passwordPepper,
		jwtSecret:      jwtSecret,
	}
}

// ============================================
// PASSWORD HASHING (Argon2id)
// ============================================

// PasswordHash contém os componentes do hash
type PasswordHash struct {
	Salt string // base64
	Hash string // base64
}

// HashPassword cria hash Argon2id com salt aleatorio e pepper
func (cs *CryptoService) HashPassword(password string) (string, error) {
	// Gerar salt aleatorio (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("erro ao gerar salt: %w", err)
	}

	// Concatenar senha + pepper
	passwordWithPepper := password + cs.passwordPepper

	// Hash com Argon2id
	// Time=3, Memory=64MB, Parallelism=4, SaltLength=16, KeyLength=32
	hash := argon2.IDKey(
		[]byte(passwordWithPepper),
		salt,
		3,       // time cost (iterations)
		64*1024, // memory cost (64 MB)
		4,       // parallelism
		32,      // key length (256 bits)
	)

	// Formato: base64(salt)$base64(hash)
	encoded := fmt.Sprintf(
		"%s$%s",
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword valida se password corresponde ao hash armazenado
func (cs *CryptoService) VerifyPassword(password, storedHash string) (bool, error) {
	// Formato esperado: base64(salt)$base64(hash)
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("formato de hash inválido")
	}

	saltStr := parts[0]
	hashStr := parts[1]

	// Decodificar salt
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false, fmt.Errorf("erro ao decodificar salt: %w", err)
	}

	// Decodificar hash armazenado
	storedHashBytes, err := base64.StdEncoding.DecodeString(hashStr)
	if err != nil {
		return false, fmt.Errorf("erro ao decodificar hash: %w", err)
	}

	// Recompor senha com pepper
	passwordWithPepper := password + cs.passwordPepper

	// Refazer hash com salt extraído usando mesmos parâmetros
	// Time=3, Memory=64MB, Parallelism=4, KeyLength=32
	newHash := argon2.IDKey(
		[]byte(passwordWithPepper),
		salt,
		3,       // time cost (iterations)
		64*1024, // memory cost (64 MB)
		4,       // parallelism
		32,      // key length (256 bits)
	)

	// Constant-time comparison (previne timing attacks)
	return constantTimeCompare(newHash, storedHashBytes), nil
}

// constantTimeCompare compara dois byte slices em tempo constante
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}

	return diff == 0
}

// ============================================
// JWT TOKENS
// ============================================

// CustomClaims contém os claims customizados do JWT
type CustomClaims struct {
	Sub                    string `json:"sub"` // user_id
	Email                  string `json:"email"`
	HasCompletedOnboarding bool   `json:"hasCompletedOnboarding"`
	OnboardingStep         int    `json:"onboardingStep"`
	jwt.RegisteredClaims
}

// GenerateAccessToken cria um access token (15 minutos)
func (cs *CryptoService) GenerateAccessToken(userID uuid.UUID, email string, hasCompletedOnboarding bool, onboardingStep int) (string, error) {
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute)

	claims := CustomClaims{
		Sub:                    userID.String(),
		Email:                  email,
		HasCompletedOnboarding: hasCompletedOnboarding,
		OnboardingStep:         onboardingStep,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
			Issuer:    "organiq-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cs.jwtSecret))
}

// ValidateAccessToken valida e extrai claims do access token
func (cs *CryptoService) ValidateAccessToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
			}
			return []byte(cs.jwtSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

// GenerateRefreshToken cria um refresh token (UUID simples)
// Este será armazenado hashed no banco
func (cs *CryptoService) GenerateRefreshToken() (string, error) {
	return uuid.New().String(), nil
}

// HashRefreshToken cria hash SHA-256 do refresh token para armazenamento
func (cs *CryptoService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyRefreshTokenHash valida refresh token contra hash armazenado
func (cs *CryptoService) VerifyRefreshTokenHash(token, storedHash string) bool {
	hash := sha256.Sum256([]byte(token))
	computed := base64.StdEncoding.EncodeToString(hash[:])
	return constantTimeCompare([]byte(computed), []byte(storedHash))
}

// ============================================
// FIELD ENCRYPTION (AES-256-GCM)
// ============================================

// EncryptAES criptografa dados sensíveis (ex: app passwords)
// Retorna: base64(iv+ciphertext)
func (cs *CryptoService) EncryptAES(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(cs.jwtSecret[:32])) // Use 32 chars do jwtSecret como key
	if err != nil {
		return "", fmt.Errorf("erro ao criar cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("erro ao criar GCM: %w", err)
	}

	// Gerar IV aleatorio (nonce)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("erro ao gerar nonce: %w", err)
	}

	// Criptografar
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Retornar base64(nonce+ciphertext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES descriptografa dados criptografados
func (cs *CryptoService) DecryptAES(encrypted string) (string, error) {
	block, err := aes.NewCipher([]byte(cs.jwtSecret[:32]))
	if err != nil {
		return "", fmt.Errorf("erro ao criar cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("erro ao criar GCM: %w", err)
	}

	// Decodificar base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar base64: %w", err)
	}

	// Extrair nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext muito curto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografar
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("erro ao descriptografar: %w", err)
	}

	return string(plaintext), nil
}
