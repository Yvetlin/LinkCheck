package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linkcheck/internal/handler"
	"github.com/linkcheck/internal/queue"
	"github.com/linkcheck/internal/service"
	"github.com/linkcheck/internal/storage"
)

const (
	defaultPort     = "8080"
	shutdownTimeout = 30 * time.Second
	workersCount    = 5
)

func main() {
	storage, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Не удалось создать хранилище: %v", err)
	}

	checker := service.NewChecker()

	taskQueue := queue.NewQueue(storage, checker, workersCount)
	taskQueue.Start()

	httpHandler := handler.NewHandler(storage, taskQueue, checker)

	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("Сервер запускается на порту %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Не удалось запустить сервер: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Останавливаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке HTTP сервера: %v", err)
	}

	if err := taskQueue.Shutdown(shutdownTimeout); err != nil {
		log.Printf("Ошибка при остановке очереди задач: %v", err)
	}

	if err := storage.Save(); err != nil {
		log.Printf("Ошибка при сохранении финального состояния: %v", err)
	}

	log.Println("Сервер остановлен")
}
