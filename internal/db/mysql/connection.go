package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/JonayMedina/api-music/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func InitDB(cfg *config.Config) (*sql.DB, error) {

	Db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.MySQLUser, cfg.MySQLPass, cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDB))
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	err = Db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	Db.SetMaxOpenConns(10)
	Db.SetMaxIdleConns(5)
	Db.SetConnMaxLifetime(time.Hour)

	return Db, nil
}
