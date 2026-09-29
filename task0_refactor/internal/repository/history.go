// Package repository містить шар доступу до даних: інтерфейс сховища історії
// операцій та його реалізацію в пам'яті. Залежить лише від models.
package repository

import (
	"sync"

	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/models"
)

// History — сховище історії виконаних операцій. Інтерфейс дозволяє підмінити
// реалізацію в пам'яті на файл чи базу даних, не змінюючи service.
type History interface {
	// Save додає запис до історії.
	Save(entry models.HistoryEntry) error
	// List повертає всі записи у порядку додавання.
	List() ([]models.HistoryEntry, error)
}

// InMemoryHistory — реалізація History, що зберігає записи в пам'яті процесу.
// Безпечна для конкурентного використання.
type InMemoryHistory struct {
	mu      sync.Mutex
	entries []models.HistoryEntry
}

// Перевірка на етапі компіляції, що InMemoryHistory реалізує History.
var _ History = (*InMemoryHistory)(nil)

// NewInMemoryHistory створює порожнє сховище в пам'яті.
func NewInMemoryHistory() *InMemoryHistory {
	return &InMemoryHistory{}
}

// Save додає запис до історії. Ця реалізація ніколи не повертає помилку,
// але сигнатура відповідає інтерфейсу History, щоб інші реалізації
// (файл, база даних) могли її повертати.
func (r *InMemoryHistory) Save(entry models.HistoryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries = append(r.entries, entry)
	return nil
}

// List повертає копію всіх записів у порядку додавання, тому виклики не
// можуть змінити внутрішній стан сховища.
func (r *InMemoryHistory) List() ([]models.HistoryEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]models.HistoryEntry, len(r.entries))
	copy(out, r.entries)
	return out, nil
}
