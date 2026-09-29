// Package handler містить транспортний шар застосунку — CLI. Handler лише
// розбирає команду користувача, викликає service та форматує відповідь;
// бізнес-логіки тут немає. Залежить від service і models, але не від
// доменних пакунків чи repository.
package handler

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/models"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/service"
)

// Коди завершення процесу.
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

const usage = `usage: app <command> [args]

commands:
  add|sub|mul|div A B   arithmetic operation over two numbers
  analyze TEXT...       count words and characters in TEXT
  demo                  run the built-in demonstration (default)
  help                  print this message
`

// operations зіставляє CLI-команди з доменними операціями.
var operations = map[string]models.Operation{
	"add": models.OpAdd,
	"sub": models.OpSubtract,
	"mul": models.OpMultiply,
	"div": models.OpDivide,
}

// CLI — обробник команд командного рядка.
type CLI struct {
	app    *service.App
	out    io.Writer
	errOut io.Writer

	// writeErr — перша помилка запису в out або errOut під час поточного
	// Run. Запис не перериває обробку команди; помилка перевіряється один
	// раз наприкінці Run (патерн "errors are values").
	writeErr error
}

// New створює CLI, що делегує роботу app і пише результати в out,
// а помилки — в errOut.
func New(app *service.App, out, errOut io.Writer) *CLI {
	return &CLI{app: app, out: out, errOut: errOut}
}

// Run виконує команду з args (без імені програми) і повертає код завершення.
// Якщо результат не вдалося записати (наприклад, stdout закрито або канал
// розірвано), Run повертає ExitError навіть для успішно виконаної команди:
// користувач результату не побачив, тож рапортувати успіх не можна.
func (c *CLI) Run(args []string) int {
	c.writeErr = nil
	code := c.dispatch(args)
	if c.writeErr == nil {
		return code
	}

	// Best effort: якщо не працює саме errOut, це повідомлення теж буде
	// втрачено, але код завершення все одно сигналізує про збій.
	c.write(c.errOut, "error: writing output: %v\n", c.writeErr)
	if code == ExitOK {
		return ExitError
	}
	return code
}

func (c *CLI) dispatch(args []string) int {
	if len(args) == 0 {
		return c.demo()
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "add", "sub", "mul", "div":
		return c.calculate(cmd, rest)
	case "analyze":
		return c.analyze(rest)
	case "demo":
		return c.demo()
	case "help", "-h", "--help":
		c.write(c.out, "%s", usage)
		return ExitOK
	default:
		c.write(c.errOut, "unknown command %q\n\n%s", cmd, usage)
		return ExitUsage
	}
}

func (c *CLI) calculate(cmd string, args []string) int {
	if len(args) != 2 {
		c.write(c.errOut, "%s expects exactly two numbers\n\n%s", cmd, usage)
		return ExitUsage
	}
	a, errA := strconv.ParseFloat(args[0], 64)
	b, errB := strconv.ParseFloat(args[1], 64)
	if err := errors.Join(errA, errB); err != nil {
		c.write(c.errOut, "%s: invalid number: %v\n", cmd, err)
		return ExitUsage
	}

	calc, err := c.app.Calculate(operations[cmd], a, b)
	if err != nil {
		return c.fail(err)
	}
	c.write(c.out, "%v\n", calc)
	return ExitOK
}

func (c *CLI) analyze(args []string) int {
	stats, err := c.app.AnalyzeText(strings.Join(args, " "))
	if err != nil {
		return c.fail(err)
	}
	c.write(c.out, "words: %d, chars: %d\n", stats.Words, stats.Chars)
	return ExitOK
}

// demo відтворює сценарій початкового main.go: успішна операція, розпізнавання
// ділення на нуль через errors.Is, підрахунок слів — і наприкінці історія
// операцій, що пройшла через усі шари (handler -> service -> repository).
func (c *CLI) demo() int {
	sum, err := c.app.Calculate(models.OpAdd, 2, 3)
	if err != nil {
		return c.fail(err)
	}
	c.write(c.out, "2 + 3 = %v\n", sum.Result)

	if _, err := c.app.Calculate(models.OpDivide, 10, 0); err != nil {
		if !errors.Is(err, service.ErrDivisionByZero) {
			return c.fail(err)
		}
		c.write(c.out, "division by zero was correctly detected: %v\n", err)
	}

	stats, err := c.app.AnalyzeText("the quick brown fox")
	if err != nil {
		return c.fail(err)
	}
	c.write(c.out, "word count: %d\n", stats.Words)

	entries, err := c.app.History()
	if err != nil {
		return c.fail(err)
	}
	c.write(c.out, "history:\n")
	for i, e := range entries {
		c.write(c.out, "  %d. [%s] %s\n", i+1, e.Kind, e.Summary)
	}
	return ExitOK
}

// fail друкує помилку у зручному для користувача вигляді. Відомі помилки
// (sentinel errors сервісу) розпізнаються через errors.Is і отримують
// зрозуміле пояснення; решта виводяться як є.
func (c *CLI) fail(err error) int {
	switch {
	case errors.Is(err, service.ErrDivisionByZero):
		c.write(c.errOut, "error: cannot divide by zero\n")
	case errors.Is(err, service.ErrNonFiniteInput):
		c.write(c.errOut, "error: operands must be finite numbers (NaN and Inf are not allowed)\n")
	case errors.Is(err, service.ErrNonFiniteResult):
		c.write(c.errOut, "error: result is not a finite number (overflow)\n")
	case errors.Is(err, service.ErrEmptyText):
		c.write(c.errOut, "error: text must not be empty\n")
	default:
		c.write(c.errOut, "error: %v\n", err)
	}
	return ExitError
}

// write форматує та записує повідомлення у w, запам'ятовуючи першу помилку
// запису для перевірки наприкінці Run.
func (c *CLI) write(w io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(w, format, args...); err != nil && c.writeErr == nil {
		c.writeErr = err
	}
}
