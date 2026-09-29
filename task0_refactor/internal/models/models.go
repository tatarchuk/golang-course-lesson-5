// Package models містить доменні типи, спільні для всіх шарів застосунку
// (handler, service, repository). Пакунок не залежить від інших пакунків
// проєкту, тому будь-який шар може імпортувати його без циклів.
package models

import "fmt"

// Operation — вид арифметичної операції.
type Operation string

// Підтримувані арифметичні операції.
const (
	OpAdd      Operation = "add"
	OpSubtract Operation = "subtract"
	OpMultiply Operation = "multiply"
	OpDivide   Operation = "divide"
)

// Calculation — виконана арифметична операція разом із результатом.
type Calculation struct {
	Op     Operation
	A      float64
	B      float64
	Result float64
}

// String повертає короткий людиночитний запис, наприклад "add(2, 3) = 5".
func (c Calculation) String() string {
	return fmt.Sprintf("%s(%v, %v) = %v", c.Op, c.A, c.B, c.Result)
}

// TextStats — статистика тексту: кількість слів і символів.
type TextStats struct {
	Words int
	Chars int
}

// Типи записів історії.
const (
	HistoryCalculation = "calculation"
	HistoryText        = "text"
)

// HistoryEntry — запис в історії виконаних операцій.
type HistoryEntry struct {
	// Kind — тип запису: HistoryCalculation або HistoryText.
	Kind string
	// Summary — людиночитний опис операції та її результату.
	Summary string
}
