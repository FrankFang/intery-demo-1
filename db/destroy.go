package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"log"
)

func Destroy(name string) error {
	pgConn, err := pgconn.Connect(context.Background(), "postgresql://intery:123456@psql1:5432/postgres")
	if err != nil {
		log.Fatalln("pgconn failed to connect: ", err)
		return err
	}
	defer pgConn.Close(context.Background())

	result := pgConn.Exec(context.Background(), "drop database "+name)
	if _, err := result.ReadAll(); err != nil {
		log.Fatalln("drop database failed: ", err)
		return err
	}
	fmt.Println("Database dropped!")
	return nil
}
