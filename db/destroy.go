package db

import (
	"context"
	"fmt"
	"intery/server/database"
	"log"

	"github.com/jackc/pgconn"
)

func Drop() error {
	host, user, name, password, port := database.GetDsn()
	pgConn, err := pgconn.Connect(
		context.Background(),
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/postgres",
			user,
			password,
			host,
			port,
		),
	)
	if err != nil {
		log.Fatalln("pgconn failed to connect: ", err)
		return err
	}
	defer pgConn.Close(context.Background())

	result := pgConn.Exec(context.Background(), "drop database if exists "+name)
	if _, err := result.ReadAll(); err != nil {
		log.Fatalf("drop database %v failed: %v \n", name, err)
		return err
	}
	fmt.Printf("Database %v dropped!\n", name)
	return nil
}
