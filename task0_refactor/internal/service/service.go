// Package service містить бізнес-логіку застосунку. Він оркеструє доменні
// пакунки calculator і textanalyzer та зберігає історію операцій через
// repository. Service нічого не знає про спосіб взаємодії з користувачем
// (CLI, HTTP тощо) — це відповідальність handler.
package service

import (
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/calculator"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/models"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/repository"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/textanalyzer"
)

// Помилки, які може повертати service. Доменні sentinel-помилки
// реекспортуються, щоб handler перевіряв їх через errors.Is, не імпортуючи
// доменні пакунки напряму.
var (
	ErrUnknownOperation = errors.New("unknown operation")
	ErrNonFiniteInput   = errors.New("operands must be finite numbers")
	ErrNonFiniteResult  = errors.New("result is not a finite number")
	ErrDivisionByZero   = calculator.ErrDivisionByZero
	ErrEmptyText        = textanalyzer.ErrEmptyText
)

// App — сервісний шар застосунку. Створюйте через New.
type App struct {
	history repository.History
	logger  *slog.Logger
}

// New створює сервіс, що зберігає історію операцій у history. logger
// отримує попередження про другорядні збої (наприклад, не вдалося зберегти
// запис історії); nil означає slog.Default().
func New(history repository.History, logger *slog.Logger) *App {
	if logger == nil {
		logger = slog.Default()
	}
	return &App{history: history, logger: logger}
}

// Calculate виконує арифметичну операцію op над a та b і записує результат
// в історію. Помилки, що повертаються (усі обгорнуті через %w):
//   - ErrNonFiniteInput — a або b є NaN чи ±Inf
//   - ErrUnknownOperation — op не є однією з підтримуваних операцій
//   - ErrDivisionByZero — ділення на нуль
//   - ErrNonFiniteResult — результат переповнився до ±Inf або NaN
func (s *App) Calculate(op models.Operation, a, b float64) (models.Calculation, error) {
	if !isFinite(a) || !isFinite(b) {
		return models.Calculation{}, fmt.Errorf("service: calculate %s(%v, %v): %w", op, a, b, ErrNonFiniteInput)
	}

	var result float64
	switch op {
	case models.OpAdd:
		result = calculator.Add(a, b)
	case models.OpSubtract:
		result = calculator.Subtract(a, b)
	case models.OpMultiply:
		result = calculator.Multiply(a, b)
	case models.OpDivide:
		quotient, err := calculator.Divide(a, b)
		if err != nil {
			return models.Calculation{}, fmt.Errorf("service: calculate: %w", err)
		}
		result = quotient
	default:
		return models.Calculation{}, fmt.Errorf("service: calculate: %w %q", ErrUnknownOperation, op)
	}

	if !isFinite(result) {
		return models.Calculation{}, fmt.Errorf("service: calculate %s(%v, %v): %w", op, a, b, ErrNonFiniteResult)
	}

	calc := models.Calculation{Op: op, A: a, B: b, Result: result}
	s.record(models.HistoryCalculation, calc.String())
	return calc, nil
}

// AnalyzeText рахує слова та символи в text і записує результат в історію.
// Порожній текст повертає помилку, що обгортає ErrEmptyText.
func (s *App) AnalyzeText(text string) (models.TextStats, error) {
	words, err := textanalyzer.WordCount(text)
	if err != nil {
		return models.TextStats{}, fmt.Errorf("service: analyze text: %w", err)
	}
	stats := models.TextStats{Words: words, Chars: textanalyzer.CharCount(text)}

	summary := fmt.Sprintf("analyze(%q): %d words, %d chars", text, stats.Words, stats.Chars)
	s.record(models.HistoryText, summary)
	return stats, nil
}

// History повертає всі успішно виконані операції у порядку виконання.
func (s *App) History() ([]models.HistoryEntry, error) {
	entries, err := s.history.List()
	if err != nil {
		return nil, fmt.Errorf("service: history: %w", err)
	}
	return entries, nil
}

// record зберігає запис в історії. Історія — другорядна функція: якщо
// зберегти запис не вдалося, операція все одно вважається успішною, а збій
// лише логується як попередження, щоб не ховати його мовчки.
func (s *App) record(kind, summary string) {
	entry := models.HistoryEntry{Kind: kind, Summary: summary}
	if err := s.history.Save(entry); err != nil {
		s.logger.Warn("history entry not saved", "kind", kind, "summary", summary, "error", err)
	}
}

// isFinite повідомляє, чи x є звичайним числом (не NaN і не ±Inf).
func isFinite(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}
