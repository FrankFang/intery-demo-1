package db

import (
	"context"
	"fmt"
	"intery/server/database"
	"log"

	"github.com/jackc/pgconn"
)

func Create() (err error) {
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

	result := pgConn.Exec(
		context.Background(),
		fmt.Sprintf(`SELECT FROM pg_database WHERE datname = '%v'`, name),
	)

	r, err := result.ReadAll()
	if err != nil {
		log.Fatalf("select database %v failed: %v \n", name, err)
		return err
	}
	if len(r[0].Rows) == 0 {
		result := pgConn.Exec(context.Background(), fmt.Sprintf(`CREATE DATABASE %v`, name))
		if _, err := result.ReadAll(); err != nil {
			log.Fatalf("create database %v failed: %v \n", name, err)
			return err
		}
		fmt.Printf("Database %v created!\n", name)
	} else {
		fmt.Printf("Datebase %v exists, skip creation\n", name)
	}
	return nil
}
