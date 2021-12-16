package db

import (
	"log"
)

func Migrate(name string) error {
	m := NewMigrate(name)
	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
		return err
	}
	log.Printf("Migration did run successfully")
	return nil
}
