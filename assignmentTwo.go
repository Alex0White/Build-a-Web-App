package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	prepareDatabase()

}
func addNote(NoteId int, username string, note string) {
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO NotesTable(noteId, username, password) VALUES($1,$2, $3)", NoteId, username, note)
	if err != nil {
		log.Fatal(err)
	}

}
func addUser(username string, password string) {
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO UsersTable(username, password) VALUES($1,$2)", username, password)
	if err != nil {
		log.Fatal(err)
	}

}

func prepareDatabase() {
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	err = db.Ping()

	//userstable
	_, err = db.Exec("DROP TABLE IF EXISTS UsersTable")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE UsersTable(username varchar(50), password varchar(50))")
	if err != nil {
		log.Fatal(err)
	}

	//notestable
	_, err = db.Exec("DROP TABLE IF EXISTS NotesTable")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE NotesTable(noteId int, username varchar(50), note varchar(50))")
	if err != nil {
		log.Fatal(err)
	}
	//permissions table
	_, err = db.Exec("DROP TABLE IF EXISTS PermissionsTable ")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE PermissionsTable(noteId int, username varchar(50), read boolean, write boolean)")
	if err != nil {
		log.Fatal(err)
	}

}

type User struct {
	username string
	password string
}
type Note struct {
	username string
	noteId   float64
	note     string
}
type Permissions struct {
	noteId   int
	username string
	read     bool
	write    bool
}
