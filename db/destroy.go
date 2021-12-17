package db

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
)

func Drop(name string) error {
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
	result := pgConn.Exec(context.Background(), "drop database if exists "+name)
	if _, err := result.ReadAll(); err != nil {
		log.Fatalf("drop database %v failed: %v \n", name, err)
		return err
	}
	fmt.Printf("Database %v dropped!\n", name)
	return nil
}
