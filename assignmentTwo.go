package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	prepareDatabase()
	//viewUsers()
	//viewNotes()
	//viewPermissions()
	//uncomment to reset database

	http.HandleFunc("/adduser", addNewUser)
	http.HandleFunc("/addnote", addNewNote)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	//viewUsers()

}
func addNote(NoteId int, username string, note string) { //adds a new note to the database
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO NotesTable(noteId, username, note) VALUES($1,$2,$3)", NoteId, username, note)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO PermissionsTable(noteId, username, read, write) VALUES($1,$2,$3,$4)", NoteId, username, "true", "true") //adds read and write permissions to the user the created the note
	if err != nil {
		log.Fatal(err)
	}

}
func addUser(username string, password string) { //adds a new user to the database
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO UsersTable(username, password) VALUES($1,$2)", username, password)
	if err != nil {
		log.Fatal(err)
	}

}
func changePermissions(noteId int, username string, read string, write string) { //needs to be change permissions, if the user is already associated with the note
	db, err := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	_, err = db.Exec("INSERT INTO PermissionsTable(noteId, username, read, write) VALUES($1,$2,$3,$4)", noteId, username, read, write)
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
func addNewUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	decoder := json.NewDecoder(r.Body)
	data := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	err := decoder.Decode(&data)

	if err != nil {
		panic(err)
	}
	addUser(data.Username, data.Password) //adds new user to the database
	//fmt.Println(data.Username)
	//fmt.Println(data.Password)
	defer r.Body.Close()

}
func addNewNote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	decoder := json.NewDecoder(r.Body)
	data := struct {
		NoteId   int    `json:"noteid"`
		Username string `json:"username"`
		Note     string `json:"note`
	}{}
	err := decoder.Decode(&data)

	if err != nil {
		panic(err)
	}
	addNote(data.NoteId, data.Username, data.Note) //adds new note to the database

	defer r.Body.Close()

}

func viewUsers() {
	db, _ := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	rows, err := db.Query("SELECT * FROM UsersTable ")
	if err != nil {
		log.Fatal(err)
	}

	var (
		Username string
		Password string
	)

	for rows.Next() {

		err = rows.Scan(&Username, &Password)
		fmt.Println(Username)
		fmt.Println(Password)
		//fmt.Println(err)

	}
}
func viewNotes() {
	db, _ := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	rows, err := db.Query("SELECT * FROM NotesTable ")
	if err != nil {
		log.Fatal(err)
	}

	var (
		NoteId   int
		Username string
		Note     string
	)

	for rows.Next() {

		err = rows.Scan(&NoteId, &Username, &Note)
		fmt.Println(Note)
		fmt.Println(Username)
		fmt.Println(Note)

	}

}
func viewPermissions() {
	db, _ := sql.Open("postgres", "user=postgres password=admin dbname=webAppDatabase sslmode=disable")
	rows, err := db.Query("SELECT * FROM PermissionsTable ")
	if err != nil {
		log.Fatal(err)
	}

	var (
		NoteId   int
		Username string
		Read     bool
		Write    bool
	)

	for rows.Next() {

		err = rows.Scan(&NoteId, &Username, &Read, &Write)
		fmt.Println(NoteId)
		fmt.Println(Username)
		fmt.Println(Read)
		fmt.Println(Write)

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
