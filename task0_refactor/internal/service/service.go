// Package service містить бізнес-логіку застосунку. Він оркеструє доменні
// пакунки calculator і textanalyzer та зберігає історію операцій через
// repository. Service нічого не знає про спосіб взаємодії з користувачем
// (CLI, HTTP тощо) — це відповідальність handler.
package service

import (
	"errors"
	"fmt"

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
	ErrDivisionByZero   = calculator.ErrDivisionByZero
	ErrEmptyText        = textanalyzer.ErrEmptyText
)

// App — сервісний шар застосунку. Створюйте через New.
type App struct {
	history repository.History
}

// New створює сервіс, що зберігає історію операцій у history.
func New(history repository.History) *App {
	return &App{history: history}
}

// Calculate виконує арифметичну операцію op над a та b і записує результат
// в історію. Невідома операція повертає помилку, що обгортає
// ErrUnknownOperation; ділення на нуль — помилку, що обгортає ErrDivisionByZero.
func (s *App) Calculate(op models.Operation, a, b float64) (models.Calculation, error) {
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

	calc := models.Calculation{Op: op, A: a, B: b, Result: result}
	if err := s.record(models.HistoryCalculation, calc.String()); err != nil {
		return models.Calculation{}, err
	}
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
	if err := s.record(models.HistoryText, summary); err != nil {
		return models.TextStats{}, err
	}
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

func (s *App) record(kind, summary string) error {
	entry := models.HistoryEntry{Kind: kind, Summary: summary}
	if err := s.history.Save(entry); err != nil {
		return fmt.Errorf("service: save history: %w", err)
	}
	return nil
}
