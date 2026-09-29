// Package retry містить завдання 2 домашньої роботи: retry-обгортку
// навколо "нестабільної" операції, що використовує sentinel error
// ErrTemporary для розпізнавання відновлюваних відмов.
package retry

import (
	"errors"
	"fmt"
	"time"
)

// ErrTemporary — sentinel error, що позначає тимчасову (відновлювану) відмову
// операції. Обгортайте ErrTemporary через %w у своїй "нестабільній" функції,
// коли відмова є тимчасовою і варта повторної спроби.
var ErrTemporary = errors.New("retry: temporary failure")

// ErrInvalidMaxAttempts — sentinel error для некоректної кількості спроб
// (maxAttempts < 1).
var ErrInvalidMaxAttempts = errors.New("retry: maxAttempts must be >= 1")

// ErrNilOperation — sentinel error для випадку, коли в Do передано nil
// замість операції.
var ErrNilOperation = errors.New("retry: nil operation")

// Operation — довільна операція, що може повернути помилку.
type Operation func() (string, error)

// Do виконує op до maxAttempts разів. Між невдалими спробами робить паузу
// тривалістю backoff. Повторює спробу лише тоді, коли помилка є тимчасовою
// (errors.Is(err, ErrTemporary)) — для інших помилок Do одразу повертає
// обгорнуту помилку без повторних спроб.
//
// Поведінка:
//   - op не може бути nil, інакше повертається ErrNilOperation
//   - maxAttempts має бути >= 1, інакше повертається помилка, що обгортає
//     ErrInvalidMaxAttempts
//   - якщо операція вдається одразу — повторів немає
//   - між спробами (окрім останньої) Do чекає backoff
//   - якщо помилка НЕ є errors.Is(err, ErrTemporary) — повторів немає,
//     одразу повертається обгорнута помилка
//   - якщо всі спроби вичерпано — повертається фінальна помилка, що через
//     errors.Join зберігає помилки ВСІХ спроб (з їх номерами) і через
//     errors.Is все ще розпізнається як ErrTemporary
func Do(op Operation, maxAttempts int, backoff time.Duration) (string, error) {
	if op == nil {
		return "", ErrNilOperation
	}
	if maxAttempts < 1 {
		return "", fmt.Errorf("%w: got %d", ErrInvalidMaxAttempts, maxAttempts)
	}

	var attemptErrs []error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := op()
		if err == nil {
			return result, nil
		}
		if !errors.Is(err, ErrTemporary) {
			return "", fmt.Errorf("retry: attempt %d/%d failed with a non-temporary error: %w", attempt, maxAttempts, err)
		}
		attemptErrs = append(attemptErrs, fmt.Errorf("attempt %d/%d: %w", attempt, maxAttempts, err))
		if attempt < maxAttempts {
			time.Sleep(backoff)
		}
	}

	// errors.Join зберігає контекст кожної невдалої спроби замість лише
	// останньої; оскільки кожна з них обгортає ErrTemporary, фінальна помилка
	// теж розпізнається через errors.Is(err, ErrTemporary).
	return "", fmt.Errorf("retry: all %d attempts failed: %w", maxAttempts, errors.Join(attemptErrs...))
}

// NewFlakyOperation — допоміжна функція для тестів/демонстрації: повертає
// Operation, що падає з ErrTemporary перші failuresBeforeSuccess разів,
// а потім повертає successValue без помилки.
func NewFlakyOperation(failuresBeforeSuccess int, successValue string) Operation {
	attempt := 0
	return func() (string, error) {
		attempt++
		if attempt <= failuresBeforeSuccess {
			return "", fmt.Errorf("flaky op: attempt %d failed: %w", attempt, ErrTemporary)
		}
		return successValue, nil
	}
}
