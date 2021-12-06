package db

import "log"

func Rollback() error {
	m := NewMigrate()
	if err := m.RollbackLast(); err != nil {
		log.Fatalf("Could not rollback: %v", err)
		return err
	}
	log.Printf("Rollback did run successfully")
	return nil
}
