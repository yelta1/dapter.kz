package crypto

import "strings"

// NormalizePhone очищает номер телефона от пробелов, скобок и приводит к формату +77XXXXXXXXX
func NormalizePhone(phone string) string {
	// 1. Оставляем только цифры
	var digits []rune
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}

	digitsStr := string(digits)
	if len(digitsStr) == 0 {
		return ""
	}

	// 2. Если начинается с 8 (например, 8708...), заменяем первую цифру на 7
	if len(digitsStr) == 11 && digitsStr[0] == '8' {
		digitsStr = "7" + digitsStr[1:]
	}

	// 3. Если это 10 цифр (например, 708...), добавляем 7 в начало
	if len(digitsStr) == 10 {
		digitsStr = "7" + digitsStr
	}

	// 4. Если длина верная (11 цифр, например 7708...), добавляем + в начало
	if len(digitsStr) == 11 {
		return "+" + digitsStr
	}

	// В противном случае возвращаем исходную строку цифр (или с + если был)
	if strings.HasPrefix(phone, "+") {
		return "+" + digitsStr
	}
	return digitsStr
}
