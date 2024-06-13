package Utilities

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func DropAllTables(db *sql.DB) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			panic(err)
		}
		_, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		if err != nil {
			panic(err)
		}
		DropTable(db, table)
		_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		if err != nil {
			panic(err)
		}
	}
}

func DropTable(db *sql.DB, tableName string) {
	_, err := db.Exec("DROP TABLE IF EXISTS " + tableName)
	if err != nil {
		panic(err)
	}
}

func InitializeConnection(dbUsername string, dbPassword string, dbName string) *sql.DB {
	CreateDatabase(dbUsername, dbPassword, dbName)
	sqlConnection := ConnectToLocalDatabase(dbUsername, dbPassword, dbName)
	sqlConnection.SetMaxOpenConns(100)
	sqlConnection.SetMaxIdleConns(100)
	return sqlConnection
}

func CreateDatabase(username string, password string, dbName string) {
	sqlCon := ConnectToLocalDatabase(username, password, "")
	if err := sqlCon.Ping(); err != nil {
		panic(err)
	}
	_, err := sqlCon.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		panic(err)
	}
	sqlCon.Close()
}

func ConnectToLocalDatabase(username string, password string, dbName string) *sql.DB {
	sqlCon, err := sql.Open("mysql", username+":"+password+"@/"+dbName+"?parseTime=true")
	if err != nil {
		panic(err)
	}
	if !PingMysqlConnection(sqlCon) {
		panic("error in PingMysqlConnection")
	}
	return sqlCon
}

func PingMysqlConnection(sqlCon *sql.DB) bool {
	if err := sqlCon.Ping(); err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
