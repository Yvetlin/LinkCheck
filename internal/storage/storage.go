package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/linkcheck/pkg/types"
)

const stateFile = "state.json"

// Storage управляет сохранением состояния приложения в файл
type Storage struct {
	mu    sync.RWMutex
	state *types.State
}

// NewStorage создает новое хранилище и загружает данные из файла
func NewStorage() (*Storage, error) {
	s := &Storage{
		state: &types.State{
			LinksSets:    make(map[int]types.LinksSet),
			NextLinksNum: 1,
			PendingTasks: make([]types.PendingTask, 0),
		},
	}

	if err := s.Load(); err != nil {
		return nil, fmt.Errorf("не удалось загрузить состояние: %w", err)
	}

	return s, nil
}

// Load загружает состояние из файла state.json
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var state types.State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("не удалось распарсить состояние: %w", err)
	}

	if state.LinksSets == nil {
		state.LinksSets = make(map[int]types.LinksSet)
	}
	if state.PendingTasks == nil {
		state.PendingTasks = make([]types.PendingTask, 0)
	}

	s.state = &state
	return nil
}

// Save сохраняет текущее состояние в файл
func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось преобразовать в JSON: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}

// GetState возвращает копию текущего состояния
func (s *Storage) GetState() *types.State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stateCopy := &types.State{
		LinksSets:    make(map[int]types.LinksSet),
		NextLinksNum: s.state.NextLinksNum,
		PendingTasks: make([]types.PendingTask, len(s.state.PendingTasks)),
	}

	for k, v := range s.state.LinksSets {
		linksCopy := make(map[string]types.LinkStatus)
		for url, status := range v.Links {
			linksCopy[url] = status
		}
		stateCopy.LinksSets[k] = types.LinksSet{
			LinksNum: v.LinksNum,
			Links:    linksCopy,
		}
	}

	copy(stateCopy.PendingTasks, s.state.PendingTasks)

	return stateCopy
}

// AddLinksSet добавляет новый набор ссылок и возвращает присвоенный номер
func (s *Storage) AddLinksSet(links map[string]types.LinkStatus) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	linksNum := s.state.NextLinksNum
	s.state.LinksSets[linksNum] = types.LinksSet{
		LinksNum: linksNum,
		Links:    links,
	}
	s.state.NextLinksNum++

	return linksNum
}

// GetLinksSets возвращает наборы ссылок по их номерам
func (s *Storage) GetLinksSets(linksNums []int) map[int]types.LinksSet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int]types.LinksSet)

	for _, num := range linksNums {
		if set, exists := s.state.LinksSets[num]; exists {
			linksCopy := make(map[string]types.LinkStatus)
			for url, status := range set.Links {
				linksCopy[url] = status
			}
			result[num] = types.LinksSet{
				LinksNum: set.LinksNum,
				Links:    linksCopy,
			}
		}
	}

	return result
}

// AddPendingTask добавляет задачу в список ожидающих обработки
func (s *Storage) AddPendingTask(task types.PendingTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state.PendingTasks = append(s.state.PendingTasks, task)
}

// RemovePendingTask удаляет задачу из списка ожидающих
func (s *Storage) RemovePendingTask(linksNum int, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.state.PendingTasks {
		if task.LinksNum == linksNum && task.URL == url {
			s.state.PendingTasks = append(
				s.state.PendingTasks[:i],
				s.state.PendingTasks[i+1:]...,
			)
			break
		}
	}
}

// GetPendingTasks возвращает список всех задач, которые ожидают обработки
func (s *Storage) GetPendingTasks() []types.PendingTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]types.PendingTask, len(s.state.PendingTasks))
	copy(tasks, s.state.PendingTasks)
	return tasks
}

// UpdateLinksSetStatus обновляет статус конкретной ссылки в наборе
func (s *Storage) UpdateLinksSetStatus(linksNum int, url string, status types.LinkStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if set, exists := s.state.LinksSets[linksNum]; exists {
		if set.Links == nil {
			set.Links = make(map[string]types.LinkStatus)
		}
		set.Links[url] = status
		s.state.LinksSets[linksNum] = set
	}
}
