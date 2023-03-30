package storage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/parMaster/rpid/config"
	"github.com/parMaster/rpid/storage/model"
	"github.com/parMaster/rpid/storage/sqlite"
)

// Storer is an interface that describes the methods that a storage backend must implement.
type Storer interface {
	// Read reads records for the given module from the database.
	Read(context.Context, string) ([]model.Data, error)
	// Write writes the data to the database.
	Write(context.Context, model.Data) error
	// View returns the data for the given module in the format that is suitable for the web view.
	View(context.Context, string) (map[string]map[string]string, error)
}

func Load(ctx context.Context, cfg config.Storage, s *Storer) error {
	var err error
	switch cfg.Type {
	case "sqlite":
		*s, err = sqlite.NewStorage(ctx, cfg.Path)
		if err != nil {
			return fmt.Errorf("failed to init SQLite storage: %e", err)
		}
	case "":
		log.Printf("[DEBUG] Storage is not configured")
		return errors.New("storage is not configured")
	default:
		return fmt.Errorf("storage type %s is not supported", cfg.Type)
	}
	return err
}
