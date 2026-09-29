// Package validation містить завдання 1 домашньої роботи: валідатор форми
// реєстрації, що повертає власний (custom) тип помилки з переліком УСІХ
// невалідних полів, а не лише першого.
package validation

import (
	"strings"
	"unicode/utf8"
)

// Назви полів форми, які потрапляють у ValidationError.Fields.
const (
	fieldEmail    = "email"
	fieldPassword = "password"
	fieldAge      = "age"
)

// Обмеження правил валідації.
const (
	minPasswordLength = 8
	minAge            = 0
	maxAge            = 150
)

// RegistrationForm — вхідні дані форми реєстрації, які потрібно перевірити.
type RegistrationForm struct {
	Email    string
	Password string
	Age      int
}

// ValidationError — власний тип помилки, що переносить структуровані дані:
// список назв усіх полів, які не пройшли валідацію.
type ValidationError struct {
	// Fields — назви невалідних полів у порядку їх оголошення у формі
	// (email, password, age); кожне поле трапляється не більше одного разу.
	Fields []string
}

// Error реалізує інтерфейс error і повертає читабельне повідомлення, що
// перелічує всі невалідні поля, наприклад:
// "registration invalid: fields email, password"
func (e *ValidationError) Error() string {
	if e == nil || len(e.Fields) == 0 {
		return "registration invalid"
	}
	return "registration invalid: fields " + strings.Join(e.Fields, ", ")
}

// ValidateRegistration перевіряє форму реєстрації та повертає
// *ValidationError з переліком ВСІХ невалідних полів, якщо форма невалідна.
// Якщо форма валідна, повертає nil.
//
// Правила валідації:
//   - Email не може бути порожнім
//   - Password не може бути порожнім і має містити щонайменше 8 символів
//   - Age має бути в межах [0, 150]
//
// Перевірка не зупиняється на першому невалідному полі — усі помилки
// збираються одразу.
func ValidateRegistration(f RegistrationForm) error {
	var fields []string

	if f.Email == "" {
		fields = append(fields, fieldEmail)
	}
	// Порожній пароль автоматично коротший за minPasswordLength, тому одна
	// перевірка покриває обидві умови. Довжина рахується в символах (рунах),
	// а не в байтах, щоб не-ASCII паролі оцінювались коректно.
	if utf8.RuneCountInString(f.Password) < minPasswordLength {
		fields = append(fields, fieldPassword)
	}
	if f.Age < minAge || f.Age > maxAge {
		fields = append(fields, fieldAge)
	}

	if len(fields) == 0 {
		// Повертаємо саме нетипізований nil: nil-вказівник *ValidationError,
		// загорнутий в інтерфейс error, не дорівнював би nil.
		return nil
	}
	return &ValidationError{Fields: fields}
}
