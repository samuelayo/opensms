package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto_EncryptDecrypt(t *testing.T) {
	// 32-byte key for AES-256
	key := "12345678901234567890123456789012"
	crypto, err := NewCrypto(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"special characters", "!@#$%^&*()_+-=[]{}|;:,.<>?"},
		{"unicode", "你好世界 🌍"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
			"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := crypto.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt
			decrypted, err := crypto.Decrypt(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestCrypto_InvalidKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"too short", "short"},
		{"too long", "this-is-a-very-long-key-that-exceeds-32-bytes-requirement"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCrypto(tt.key)
			assert.Error(t, err)
		})
	}
}

func TestCrypto_DecryptInvalid(t *testing.T) {
	key := "12345678901234567890123456789012"
	crypto, err := NewCrypto(key)
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext string
	}{
		{"invalid base64", "not-valid-base64!!!"},
		{"too short", "YWJj"}, // "abc" in base64
		{"random data", "SGVsbG8gV29ybGQh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := crypto.Decrypt(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "MySecurePassword123!"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Hash should be different each time (bcrypt uses random salt)
	hash2, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestVerifyPassword(t *testing.T) {
	password := "MySecurePassword123!"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", password, true},
		{"wrong password", "WrongPassword", false},
		{"empty password", "", false},
		{"case sensitive", "mysecurepassword123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(tt.password, hash)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	lengths := []int{16, 32, 64, 128}

	for _, length := range lengths {
		t.Run(string(rune(length)), func(t *testing.T) {
			token, err := GenerateSecureToken(length)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Token should be different each time
			token2, err := GenerateSecureToken(length)
			require.NoError(t, err)
			assert.NotEqual(t, token, token2)
		})
	}
}

func TestHashSHA256(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := HashSHA256(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name          string
		password      string
		minLength     int
		requireStrong bool
		wantErr       bool
	}{
		{
			name:          "strong password",
			password:      "MySecureP@ssw0rd",
			minLength:     12,
			requireStrong: true,
			wantErr:       false,
		},
		{
			name:          "too short",
			password:      "Short1!",
			minLength:     12,
			requireStrong: true,
			wantErr:       true,
		},
		{
			name:          "no uppercase",
			password:      "mysecurep@ssw0rd",
			minLength:     12,
			requireStrong: true,
			wantErr:       true,
		},
		{
			name:          "no lowercase",
			password:      "MYSECUREP@SSW0RD",
			minLength:     12,
			requireStrong: true,
			wantErr:       true,
		},
		{
			name:          "no number",
			password:      "MySecureP@ssword",
			minLength:     12,
			requireStrong: true,
			wantErr:       true,
		},
		{
			name:          "no special char",
			password:      "MySecurePassw0rd",
			minLength:     12,
			requireStrong: true,
			wantErr:       true,
		},
		{
			name:          "weak but allowed when not required",
			password:      "simplepassword",
			minLength:     12,
			requireStrong: false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password, tt.minLength, tt.requireStrong)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeriveKey(t *testing.T) {
	password := "mypassword"
	salt := "randomsalt"

	// Same password and salt should produce same key
	key1 := DeriveKey(password, salt, 32)
	key2 := DeriveKey(password, salt, 32)
	assert.Equal(t, key1, key2)

	// Different salt should produce different key
	key3 := DeriveKey(password, "differentsalt", 32)
	assert.NotEqual(t, key1, key3)

	// Different password should produce different key
	key4 := DeriveKey("differentpassword", salt, 32)
	assert.NotEqual(t, key1, key4)

	// Check key length
	assert.Equal(t, 32, len(key1))
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "MySecurePassword123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "MySecurePassword123!"
	hash, _ := HashPassword(password)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	crypto, _ := NewCrypto(key)
	plaintext := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.Encrypt(plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	crypto, _ := NewCrypto(key)
	ciphertext, _ := crypto.Encrypt("hello world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.Decrypt(ciphertext)
	}
}
