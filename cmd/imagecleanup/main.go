package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	var before int
	if err := db.QueryRow("select count(*) from images").Scan(&before); err != nil {
		log.Fatalf("count before delete: %v", err)
	}
	fmt.Printf("== images before ==\n%d\n", before)

	result, err := db.Exec("delete from images")
	if err != nil {
		log.Fatalf("delete images: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("rows affected: %v", err)
	}
	fmt.Printf("== deleted images ==\n%d\n", affected)

	var after int
	if err := db.QueryRow("select count(*) from images").Scan(&after); err != nil {
		log.Fatalf("count after delete: %v", err)
	}
	fmt.Printf("== images after ==\n%d\n", after)
}
