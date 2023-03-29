package storage

import "context"

// Describes interface Storer, which is used to store data in the database.

// Storer is an interface that describes the methods that a storage
// backend must implement.
type Storer interface {

	// Read reads records for the given module from the database.
	Read(context.Context, string) ([]Data, error)
	// Write writes the data to the database.
	Write(context.Context, Data) error
	// Close closes the database connection.
	Close()
}

type Data struct {
	Module   string
	DateTime string
	Topic    string
	Value    string
}
