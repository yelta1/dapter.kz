package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// ErrInvalidCiphertext возвращается при невозможности расшифровать текст
var ErrInvalidCiphertext = errors.New("некорректный зашифрованный текст или ключ")

// Encrypt шифрует строку plainText с использованием AES-256-GCM.
// keyHex должен быть 32-байтовым (64 символа) ключом в hex-формате.
// Возвращает зашифрованную строку в формате Base64.
func Encrypt(plainText string, keyHex string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования ключа: %w", err)
	}

	if len(key) != 32 {
		return "", fmt.Errorf("неверная длина ключа шифрования (должна быть 32 байта, получено %d)", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("ошибка создания шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("ошибка создания GCM: %w", err)
	}

	// Создаем случайный инициализирующий вектор (nonce)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("ошибка генерации nonce: %w", err)
	}

	// Шифруем данные. Nonce записывается в начало зашифрованного слайса байт.
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Кодируем результат в Base64 для безопасного хранения в БД в виде текста
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt расшифровывает строку base64CipherText, зашифрованную с помощью AES-256-GCM.
// keyHex должен быть 32-байтовым (64 символа) ключом в hex-формате.
func Decrypt(base64CipherText string, keyHex string) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования ключа: %w", err)
	}

	if len(key) != 32 {
		return "", fmt.Errorf("неверная длина ключа дешифрования (должна быть 32 байта, получено %d)", len(key))
	}

	cipherText, err := base64.StdEncoding.DecodeString(base64CipherText)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования Base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("ошибка создания шифра: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("ошибка создания GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// Разделяем nonce и сам зашифрованный текст
	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plainTextBytes, err := gcm.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plainTextBytes), nil
}

// HashSha256 создает хэш SHA-256 от переданной строки (например, ИИН) с добавлением соли.
// Используется для уникального поиска без расшифрования данных.
func HashSha256(data string, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}
