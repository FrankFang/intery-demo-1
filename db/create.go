package db

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
)

func Create(name string) (err error) {
	pgConn, err := pgconn.Connect(context.Background(), "postgresql://intery:123456@psql1:5432/postgres")
	if err != nil {
		log.Fatalln("pgconn failed to connect: ", err)
		return err
	}
	defer pgConn.Close(context.Background())

	if gin.Mode() == gin.TestMode {
		name = fmt.Sprintf("%s_test", name)
	} else if gin.Mode() == gin.DebugMode {
		name = fmt.Sprintf("%s_development", name)
	} else {
		name = fmt.Sprintf("%s_production", name)
	}
	result := pgConn.Exec(context.Background(), fmt.Sprintf(`SELECT FROM pg_database WHERE datname = '%v'`, name))

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
