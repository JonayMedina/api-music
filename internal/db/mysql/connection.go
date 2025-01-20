package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/JonayMedina/api-music/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB(cfg *config.Config) (*sql.DB, error) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.MySQLUser, cfg.MySQLPass, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB))
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}
