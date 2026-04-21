package engines

// Engine interface defines the method each storage engine must implement
type Engine interface {
	// Set stores a key-value pair in the storage engine
	Set(key string, value string) error

	// Get retrieves a value given a key
	Get(key string) (string, error)

	// Delete removes a key and its corresponding value from the storage
	Delete(key string) error
}
