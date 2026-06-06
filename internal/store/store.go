package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	// ErrNotFound は指定したエージェントが存在しないことを表す。
	ErrNotFound = errors.New("agent not found")
	// ErrAlreadyExists は同名のエージェントが既に存在することを表す。
	ErrAlreadyExists = errors.New("agent already exists")
)

// Agent は登録済みエージェント。
type Agent struct {
	Name string `json:"name"`
	// State は json.RawMessage で保持する。エージェントが送る JSON の構造は
	// daemon 側では決めないため、struct や map へのパースは行わず、
	// 形式の検証（json.Valid）とそのままの保存・転送だけを行う。
	State json.RawMessage `json:"state"`
}

// Store はエージェント名単位の JSON 状態をインメモリで保持する。
type Store struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

// New は空の Store を返す。
func New() *Store {
	return &Store{
		agents: make(map[string]*Agent),
	}
}

// Register は name でエージェントを登録する。
func (s *Store) Register(name string) (Agent, error) {
	if name == "" {
		return Agent{}, fmt.Errorf("agent name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.agents[name]; ok {
		return Agent{}, ErrAlreadyExists
	}

	agent := &Agent{
		Name:  name,
		State: json.RawMessage(`{}`),
	}
	s.agents[name] = agent

	return cloneAgent(agent), nil
}

// List は登録済みエージェント一覧を返す。
func (s *Store) List() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]Agent, 0, len(s.agents))
	for _, agent := range s.agents {
		entries = append(entries, cloneAgent(agent))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// GetState は指定エージェントの状態 JSON を返す。
func (s *Store) GetState(name string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, ok := s.agents[name]
	if !ok {
		return nil, ErrNotFound
	}

	return append(json.RawMessage(nil), agent.State...), nil
}

// PutState は登録済みエージェントの状態を更新する。
func (s *Store) PutState(name string, state json.RawMessage) error {
	if !json.Valid(state) {
		return fmt.Errorf("invalid JSON state")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.agents[name]
	if !ok {
		return ErrNotFound
	}

	agent.State = append(json.RawMessage(nil), state...)
	return nil
}

func cloneAgent(agent *Agent) Agent {
	return Agent{
		Name:  agent.Name,
		State: append(json.RawMessage(nil), agent.State...),
	}
}
