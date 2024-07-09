package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/markdessain/go-duckdb"
)

type user struct {
	name    string
	age     int
	height  float32
	awesome bool
	bday    time.Time
}

func main() {

	conn, err := duckdb.NewConnector("", nil)
	if err != nil {
		log.Fatal(err)
	}

	db := sql.OpenDB(conn)

	_, err = db.Exec("INSTALL https; LOAD https; PRAGMA enable_progress_bar; PRAGMA disable_print_progress_bar;")
	if err != nil {
		log.Fatal(err)
	}

	check(db.Ping())

	setting := db.QueryRowContext(context.Background(), "SELECT current_setting('access_mode')")
	var am string
	check(setting.Scan(&am))
	log.Printf("DB opened with access mode %s", am)

	check(db.ExecContext(context.Background(), "CREATE TABLE users(name VARCHAR, age INTEGER, height FLOAT, awesome BOOLEAN, bday DATE)"))
	check(db.ExecContext(context.Background(), "INSERT INTO users VALUES('marc', 99, 1.91, true, '1970-01-01')"))
	check(db.ExecContext(context.Background(), "INSERT INTO users VALUES('macgyver', 70, 1.85, true, '1951-01-23')"))

	rows, err := db.QueryContext(
		context.Background(), `
		SELECT name, age, height, awesome, bday
		FROM users
		WHERE (name = ? OR name = ?) AND age > ? AND awesome = ?`,
		"macgyver", "marc", 30, true,
	)
	check(err)
	defer rows.Close()

	for rows.Next() {
		u := new(user)
		err := rows.Scan(&u.name, &u.age, &u.height, &u.awesome, &u.bday)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf(
			"%s is %d years old, %.2f tall, bday on %s and has awesomeness: %t\n",
			u.name, u.age, u.height, u.bday.Format(time.RFC3339), u.awesome,
		)
	}
	check(rows.Err())

	res, err := db.ExecContext(context.Background(), "DELETE FROM users")
	check(err)

	ra, _ := res.RowsAffected()
	log.Printf("Deleted %d rows\n", ra)

	runTransaction(db)
	testPreparedStmt(db)
	progressCheck(conn, db)
}

func check(args ...interface{}) {
	err := args[len(args)-1]
	if err != nil {
		panic(err)
	}
}

func runTransaction(db *sql.DB) {
	log.Println("Starting transaction...")
	tx, err := db.Begin()
	check(err)

	check(
		tx.ExecContext(
			context.Background(),
			"INSERT INTO users VALUES('gru', 25, 1.35, false, '1996-04-03')",
		),
	)
	row := tx.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users WHERE name = ?", "gru")
	var count int64
	check(row.Scan(&count))
	if count > 0 {
		log.Println("User Gru was inserted")
	}

	log.Println("Rolling back transaction...")
	check(tx.Rollback())

	row = db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users WHERE name = ?", "gru")
	check(row.Scan(&count))
	if count > 0 {
		log.Println("Found user Gru")
	} else {
		log.Println("Couldn't find user Gru")
	}
}

func testPreparedStmt(db *sql.DB) {
	stmt, err := db.PrepareContext(context.Background(), "INSERT INTO users VALUES(?, ?, ?, ?, ?)")
	check(err)
	defer stmt.Close()

	check(stmt.ExecContext(context.Background(), "Kevin", 11, 0.55, true, "2013-07-06"))
	check(stmt.ExecContext(context.Background(), "Bob", 12, 0.73, true, "2012-11-04"))
	check(stmt.ExecContext(context.Background(), "Stuart", 13, 0.66, true, "2014-02-12"))

	stmt, err = db.PrepareContext(context.Background(), "SELECT * FROM users WHERE age > ?")
	check(err)

	rows, err := stmt.QueryContext(context.Background(), 1)
	check(err)
	defer rows.Close()

	for rows.Next() {
		u := new(user)
		err := rows.Scan(&u.name, &u.age, &u.height, &u.awesome, &u.bday)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf(
			"%s is %d years old, %.2f tall, bday on %s and has awesomeness: %t\n",
			u.name, u.age, u.height, u.bday.Format(time.RFC3339), u.awesome,
		)
	}
}

func progressCheck(conn *duckdb.Connector, db *sql.DB) {

	go func() {
		_, err := db.Query("SELECT * FROM read_parquet('https://github.com/plotly/datasets/raw/master/2015_flights.parquet');")
		if err != nil {
			fmt.Println(err)
		}

		os.Exit(0)
	}()

	for {
		fmt.Println(conn.Progress())
		time.Sleep(time.Millisecond * 1)
	}

}
