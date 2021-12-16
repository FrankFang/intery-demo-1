package db

import "log"

func Rollback(name string) error {
	m := NewMigrate(name)
	if err := m.RollbackLast(); err != nil {
		log.Fatalf("Could not rollback: %v", err)
		return err
	}
	log.Printf("Rollback did run successfully")
	return nil
}
