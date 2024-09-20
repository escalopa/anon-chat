package storage

import (
	"encoding/json"
	"os"
	"sync"

	log "github.com/catalystgo/logger/cli"
	"github.com/escalopa/anon-chat-app/domain"
)

type InMemoryStorage struct {
	messages []domain.Message
	mu       sync.RWMutex
}

func New() *InMemoryStorage {
	return &InMemoryStorage{
		messages: make([]domain.Message, 0),
	}
}

// Load loads the messages from the file
func (s *InMemoryStorage) Load(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		// If the file does not exist, create it
		if os.IsNotExist(err) {
			log.Warnf("File %s does not exist, creating it", path)

			if errWrite := os.WriteFile(path, []byte{}, 0644); errWrite != nil {
				return errWrite
			}

			return nil
		}

		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &s.messages)
}

// Dump saves the messages to the file
func (s *InMemoryStorage) Dump(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(s.messages)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Count returns the number of stored messages
func (s *InMemoryStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.messages)
}

// GetAll returns all the stored messages
func (s *InMemoryStorage) GetAll() []domain.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.messages
}

// Add adds a new message to the storage
func (s *InMemoryStorage) Add(m domain.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, m)
}
