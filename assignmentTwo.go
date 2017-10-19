package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "sync"

	_ "github.com/lib/pq"
)

func main() {

	// Then, initialize the session manager

	prepareDatabase()
	//viewUsers()
	//viewNotes()
	//viewPermissions()
	//http.HandleFunc("/", HomePage)
	//http.HandleFunc("/", login)
	// var globalSessions *session.Manager
	// // Then, initialize the session manager

	//     globalSessions = NewManager("memory","gosessionid",3600)

	http.HandleFunc("/", login)
	http.HandleFunc("/adduser", addNewUser)
	http.HandleFunc("/notes", viewNotes)
	http.HandleFunc("/createnote", addNewNote)
	http.HandleFunc("/search", searchNotes)
	http.HandleFunc("/changepermissions", changeNewPermissions)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	//viewUsers()

}

func HomePage(w http.ResponseWriter, r *http.Request) {

	// t, err := template.ParseFiles("homepage.html") //parse the html file homepage.html

	// if err != nil { // if there is an error
	// 	log.Print("template parsing error: ", err) // log it
	// }
	// type PageVariables struct {
	// 	Test string
	// }
	// HomePageVars := PageVariables{ //store the date and time in a struct
	// 	Test: "hey",
	// }
	// err = t.Execute(w, HomePageVars) //execute the template and pass it the HomePageVars struct to fill in the gaps
	// if err != nil {                  // if there is an error
	// 	log.Print("template executing error: ", err) //log it
	// }
}

func login(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("method:", r.Method) //get request method

	if r.Method == "GET" {

		//fmt.Println(sess)

		//userName := sess.CAttr("UserName")
		//fmt.Println(userName)
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()

		// fmt.Println("username:", r.Form["username"])
		// fmt.Println("password:", r.Form["password"])

		db, _ := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable port=5432 ")
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
			if r.Form["username"][0] == Username {
				if r.Form["password"][0] == Password {
					fmt.Println("Logged in!")
					//sess := session.NewSession()
					cookie1 := &http.Cookie{Name: "username", Value: (Username), HttpOnly: false}
					http.SetCookie(w, cookie1)
					var cookie, err = r.Cookie("username")
					if err == nil {
						fmt.Println(cookie.Value)

					} else {
						fmt.Println(err)
					}

					//sess.Set("username", r.Form["username"])

					t, _ := template.ParseFiles("home.html")
					t.Execute(w, nil)
					break

				} else {
					fmt.Println("Incorrect Password!")
					break
				}
			}
		}

	}
}

func addNote(username string, note string) { //adds a new note to the database
	db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO NotesTable(username, note) VALUES($1,$2)", username, note)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO PermissionsTable(username, read, write, owner) VALUES($1,$2,$3,$4)", username, "true", "true", "true") //adds read and write permissions to the user the created the note
	if err != nil {
		log.Fatal(err)
	}

}
func addUser(username string, password string) { //adds a new user to the database

	db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	//check if username is already taken
	_, err = db.Exec("INSERT INTO UsersTable(username, password) VALUES($1,$2)", username, password)
	if err != nil {
		log.Fatal(err)
	}
	viewUsers()

}

func changePermissions(noteId int, username string, read bool, write bool, owner bool) { //needs to be change permissions, if the user is already associated with the note
	updated := false
	db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	rows, err := db.Query("SELECT * FROM PermissionsTable ")
	if err != nil {
		log.Fatal(err)
	}

	var (
		NoteId   int
		Username string
		Read     bool
		Write    bool
		Owner    bool
	)

	for rows.Next() {

		err = rows.Scan(&noteId, &username, &Read, &Write, &Owner)
		if noteId == NoteId && username == Username { //if user already has permissions associated with it it needs to be updated rather than inserted
			//update table
			updated = true

		}

	}
	if !updated {
		_, err = db.Exec("INSERT INTO PermissionsTable(noteId, username, read, write, owner) VALUES($1,$2,$3,$4,$5)", noteId, username, read, write, owner)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func prepareDatabase() {
	db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable port=5432")
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
	_, err = db.Exec("CREATE TABLE NotesTable(noteId SERIAL, username varchar(50), note varchar(50))")
	if err != nil {
		log.Fatal(err)
	}
	//permissions table
	_, err = db.Exec("DROP TABLE IF EXISTS PermissionsTable ")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE PermissionsTable(noteId int, username varchar(50), read boolean, write boolean, owner boolean)")
	if err != nil {
		log.Fatal(err)
	}

}
func addNewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("addaccount.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()

		addUser(r.Form["username"][0], r.Form["password"][0])
		t, _ := template.ParseFiles("homepage.html")
		t.Execute(w, nil)

		//fmt.Println("password:", r.Form["password"])

		//defer r.Body.Close()

	}
}
func searchNotes(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	//fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("search.html")

		t.Execute(w, nil)
	}
}

func addNewNote(w http.ResponseWriter, r *http.Request) {
	var username, err = r.Cookie("username")
	if err == nil {
		fmt.Println(username.Value)

	} else {
		fmt.Println(err)
	}

	//ess := session.Get(r)
	r.ParseForm()
	//fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {

		t, _ := template.ParseFiles("createnote.html")
		t.Execute(w, nil)
	} else {
		addNote(username.Value, r.Form["note"][0])

		t, _ := template.ParseFiles("createnote.html")
		t.Execute(w, nil)
		//viewNotes()
		viewPermissions()
	}

	// decoder := json.NewDecoder(r.Body)
	// data := struct {
	// 	NoteId   int    `json:"noteid"`
	// 	Username string `json:"username"`
	// 	Note     string `json:"note`
	// }{}
	// err := decoder.Decode(&data)

	// if err != nil {
	// 	panic(err)
	// }
	// addNote(data.NoteId, data.Username, data.Note) //adds new note to the database

	// defer r.Body.Close()

}
func changeNewPermissions(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	decoder := json.NewDecoder(r.Body)
	data := struct {
		NoteId   int    `json:"noteid"`
		Username string `json:"username"`
		Read     bool   `json:"read"`
		Write    bool   `json:"write"`
		Owner    bool   `json:"owner"`
	}{}
	err := decoder.Decode(&data)

	if err != nil {
		panic(err)
	}
	changePermissions(data.NoteId, data.Username, data.Read, data.Write, data.Owner)

	defer r.Body.Close()

}

func viewUsers() {
	db, _ := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable port=5432")
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
func viewNotes(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("notes.html")

		t.Execute(w, nil)
	}

	db, _ := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
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
		fmt.Println(NoteId)
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, Note+"<br>")

	}

}
func viewPermissions() {
	db, _ := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	rows, err := db.Query("SELECT * FROM PermissionsTable ")
	if err != nil {
		log.Fatal(err)
	}

	var (
		NoteId   int
		Username string
		Read     bool
		Write    bool
		Owner    bool
	)

	for rows.Next() {

		err = rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)
		fmt.Println(NoteId)
		fmt.Println(Username)
		fmt.Println(Read)
		fmt.Println(Write)
		fmt.Println(Owner)

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
	owner    bool
}
