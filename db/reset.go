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
	result := pgConn.Exec(context.Background(), "create database "+name)
	if _, err := result.ReadAll(); err != nil {
		log.Fatalf("create database %v failed: %v \n", name, err)
		return err
	}
	fmt.Printf("Database %v created!\n", name)
	return nil
}
