package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type SQLItem struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SQL         string `json:"sql"`
}

type SQLConfig struct {
	Description string    `json:"description"`
	SQLs        []SQLItem `json:"sqls"`
}

type Scheduler struct {
	sqlMap map[string]SQLItem
	mu     sync.RWMutex
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		sqlMap: make(map[string]SQLItem),
	}
}

func (s *Scheduler) LoadFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	return s.LoadFromJSON(data)
}

func (s *Scheduler) LoadFromJSON(data []byte) error {
	var config SQLConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range config.SQLs {
		s.sqlMap[item.UUID] = item
	}

	return nil
}

func (s *Scheduler) GetSQL(uuid string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.sqlMap[uuid]
	if !exists {
		return "", false
	}

	return item.SQL, true
}

func (s *Scheduler) GetSQLItem(uuid string) (SQLItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.sqlMap[uuid]
	return item, exists
}

func (s *Scheduler) ListAllUUIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uuids := make([]string, 0, len(s.sqlMap))
	for uuid := range s.sqlMap {
		uuids = append(uuids, uuid)
	}

	return uuids
}

func (s *Scheduler) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sqlMap)
}
