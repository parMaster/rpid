package storage

import "context"

// Storer is an interface that describes the methods that a storage backend must implement.
type Storer interface {
	// Read reads records for the given module from the database.
	Read(context.Context, string) ([]Data, error)
	// Write writes the data to the database.
	Write(context.Context, Data) error
}

type Data struct {
	Module   string
	DateTime string
	Topic    string
	Value    string
}
