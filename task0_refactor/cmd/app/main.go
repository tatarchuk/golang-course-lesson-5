// Command app — тонкий main.go, що лише "з'єднує" (wiring) шари застосунку:
// repository -> service -> handler. Уся логіка живе в internal/.
package main

import (
	"log/slog"
	"os"

	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/handler"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/repository"
	"github.com/softserve/go-with-genai-topic4-error-handling/task0_refactor/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	history := repository.NewInMemoryHistory()
	app := service.New(history, logger)
	cli := handler.New(app, os.Stdout, os.Stderr)

	os.Exit(cli.Run(os.Args[1:]))
}
