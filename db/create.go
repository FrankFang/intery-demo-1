package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgconn"
)

func Create(name string) (err error) {
	pgConn, err := pgconn.Connect(context.Background(), "postgresql://intery:123456@psql1:5432/postgres")
	if err != nil {
		log.Fatalln("pgconn failed to connect: ", err)
		return err
	}
	defer pgConn.Close(context.Background())

	result := pgConn.Exec(context.Background(), "create database "+name)
	if _, err := result.ReadAll(); err != nil {
		log.Fatalln("create database failed: ", err)
		return err
	}
	fmt.Println("Database created!")
	return nil
}
