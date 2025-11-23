package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/linkcheck/internal/service"
	"github.com/linkcheck/internal/storage"
	"github.com/linkcheck/pkg/types"
)

// Queue управляет очередью задач проверки ссылок
type Queue struct {
	storage        *storage.Storage
	checker        *service.Checker
	workers        int
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	taskChan       chan Task
	mu             sync.Mutex
	isShuttingDown bool
}

// Task представляет задачу проверки ссылки
type Task struct {
	LinksNum int
	URL      string
}

// NewQueue создает новую очередь задач
func NewQueue(storage *storage.Storage, checker *service.Checker, workers int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		storage:  storage,
		checker:  checker,
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
		taskChan: make(chan Task, 100),
	}
}

// Start запускает воркеров для обработки задач
func (q *Queue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	q.restorePendingTasks()
}

// restorePendingTasks восстанавливает задачи из сохраненного состояния
func (q *Queue) restorePendingTasks() {
	pendingTasks := q.storage.GetPendingTasks()

	for _, task := range pendingTasks {
		if task.Status == "pending" {
			select {
			case q.taskChan <- Task{LinksNum: task.LinksNum, URL: task.URL}:
			default:
			}
		}
	}
}

// worker обрабатывает задачи из очереди
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case task, ok := <-q.taskChan:
			if !ok {
				return
			}

			q.updateTaskStatus(task.LinksNum, task.URL, "processing")

			status := q.checker.CheckLink(q.ctx, task.URL)

			q.storage.UpdateLinksSetStatus(task.LinksNum, task.URL, status)

			q.storage.RemovePendingTask(task.LinksNum, task.URL)

			if err := q.storage.Save(); err != nil {
				fmt.Printf("Ошибка при сохранении состояния: %v\n", err)
			}
		}
	}
}

// updateTaskStatus обновляет статус задачи в хранилище
func (q *Queue) updateTaskStatus(linksNum int, url string, status string) {
	pendingTasks := q.storage.GetPendingTasks()
	for i, task := range pendingTasks {
		if task.LinksNum == linksNum && task.URL == url {
			pendingTasks[i].Status = status
			break
		}
	}
}

// Enqueue добавляет задачу в очередь для обработки
func (q *Queue) Enqueue(linksNum int, url string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isShuttingDown {
		return fmt.Errorf("очередь останавливается")
	}

	q.storage.AddPendingTask(types.PendingTask{
		LinksNum: linksNum,
		URL:      url,
		Status:   "pending",
	})

	if err := q.storage.Save(); err != nil {
		return fmt.Errorf("не удалось сохранить состояние: %w", err)
	}

	select {
	case q.taskChan <- Task{LinksNum: linksNum, URL: url}:
		return nil
	case <-time.After(1 * time.Second):
		return nil
	}
}

// Shutdown корректно останавливает очередь
func (q *Queue) Shutdown(timeout time.Duration) error {
	q.mu.Lock()
	q.isShuttingDown = true
	q.mu.Unlock()

	close(q.taskChan)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-shutdownCtx.Done():
		q.cancel()
		return fmt.Errorf("таймаут остановки: некоторые задачи могут быть не завершены")
	}
}

// GetPendingCount возвращает количество задач, которые ожидают обработки
func (q *Queue) GetPendingCount() int {
	return len(q.storage.GetPendingTasks())
}
