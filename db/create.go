package db

import (
	"context"
	"log"

	"github.com/jackc/pgconn"
)

func CreateDB(name string) (err error) {
	pgConn, err := pgconn.Connect(context.Background(), "postgresql://intery:123456@psql1:5432/")
	if err != nil {
		log.Fatalln("pgconn failed to connect:", err)
	}
	defer pgConn.Close(context.Background())

	result := pgConn.Exec(context.Background(), "create database "+name)
	if _, err := result.ReadAll(); err != nil {
		return err
	} else {
		return nil
	}
}
