package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// 32-байтовый ключ в hex
	keyHex := "6368616e676520746869732070617373776f726420746f203332206279746573"
	plainText := "123456789012" // Пример ИИН

	encrypted, err := Encrypt(plainText, keyHex)
	if err != nil {
		t.Fatalf("Ошибка шифрования: %v", err)
	}

	if encrypted == plainText {
		t.Error("Зашифрованный текст совпадает с исходным")
	}

	decrypted, err := Decrypt(encrypted, keyHex)
	if err != nil {
		t.Fatalf("Ошибка расшифрования: %v", err)
	}

	if decrypted != plainText {
		t.Errorf("Ожидалось получить %q, получено %q", plainText, decrypted)
	}
}

func TestDecryptWithInvalidKey(t *testing.T) {
	keyHex1 := "6368616e676520746869732070617373776f726420746f203332206279746573"
	keyHex2 := "3232323232323232323232323232323232323232323232323232323232323232"
	plainText := "ФИО Тест"

	encrypted, err := Encrypt(plainText, keyHex1)
	if err != nil {
		t.Fatalf("Ошибка шифрования: %v", err)
	}

	_, err = Decrypt(encrypted, keyHex2)
	if err == nil {
		t.Error("Ожидалась ошибка расшифрования с неверным ключом, но операция прошла успешно")
	}
}

func TestHashSha256(t *testing.T) {
	data := "123456789012"
	salt := "test_jwt_secret_salt"

	hash1 := HashSha256(data, salt)
	hash2 := HashSha256(data, salt)

	if hash1 != hash2 {
		t.Error("Хэширование должно быть детерминированным (одинаковые хэши для одинаковых входных данных)")
	}

	hashDifferent := HashSha256(data, "another_salt")
	if hash1 == hashDifferent {
		t.Error("Разные соли должны давать разные хэши для одинаковых данных")
	}
}
