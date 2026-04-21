package engines

import "fmt"

type memoryEngine struct {
	data map[string]string
}

func NewMemoryEngine() *memoryEngine {
	return &memoryEngine{data: make(map[string]string)}
}

func (m *memoryEngine) Set(key string, value string) error {
	m.data[key] = value
	return nil
}

func (m *memoryEngine) Get(key string) (value string, err error) {
	value, ok := m.data[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return value, nil

}

func (m *memoryEngine) Delete(key string) error {
	delete(m.data, key)
	return nil
}
