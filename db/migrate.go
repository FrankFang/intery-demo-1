package db

import (
	"log"
)

func Migrate() error {
	m := NewMigrate()
	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
		return err
	}
	log.Printf("Migration did run successfully")
	return nil
}
