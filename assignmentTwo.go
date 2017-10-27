package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	_ "sync"

	"regexp"

	_ "github.com/lib/pq"
)

func main() {

	prepareDatabase() //uncomment to restart the database

	//changePermissions(2,"con",false,false,true)
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

}

func login(w http.ResponseWriter, r *http.Request) {
	var loggedin = false

	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()

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
					loggedin = true
					fmt.Println("Logged in!")
					cookie1 := &http.Cookie{Name: "username", Value: (Username), HttpOnly: false}
					http.SetCookie(w, cookie1)
					var cookie, err = r.Cookie("username")
					if err == nil {
						fmt.Println(cookie.Value)

					} else {
						fmt.Println(err)
					}

					t, _ := template.ParseFiles("notes.html")
					t.Execute(w, nil)

				} else {
					fmt.Println("Incorrect Password!")

				}
			}

		}
		if !loggedin {
			fmt.Println("failed")
			t, _ := template.ParseFiles("login.html")
			t.Execute(w, nil)
		}

	}
}

func addNote(username string, note string) { //adds a new note to the database
	db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	//check if username is already taken

	query := "INSERT INTO NotesTable(username, note) VALUES($1,$2) RETURNING noteId" //returns the noteId so can be used when adding the note to the permissions table
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var NoteId int
	err = stmt.QueryRow(username, note).Scan(&NoteId)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(NoteId)
	_, err = db.Exec("INSERT INTO PermissionsTable(noteid, username, read, write, owner) VALUES($1,$2,$3,$4,$5)", NoteId, username, "true", "true", "true") //adds read and write permissions to the user the created the note
	if err != nil {
		log.Fatal(err)
	}

}
func updateNote(note string, noteid int, db *sql.DB) {

	db.Exec("UPDATE NotesTable SET note = $1 WHERE noteid = $2", note, noteid)
}
func deleteNote(username string, noteid int, db *sql.DB) {

	db.Exec("DELETE FROM NotesTable WHERE noteid = $1 AND username = $2", noteid, username)
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

		err = rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)

		if noteId == NoteId && username == Username { //if user already has permissions associated with it it needs to be updated rather than inserted
			sqlStatement := `  
			UPDATE PermissionsTable  
			SET read = $3, write = $4  
			WHERE noteid = $1 AND username = $2;`
			_, err = db.Exec(sqlStatement, noteId, username, read, write)
			if err != nil {
				panic(err)
			}
			updated = true

		}

	}
	if !updated {
		fmt.Println("wat up")
		_, err = db.Exec("INSERT INTO PermissionsTable(noteid, username, read, write, owner) VALUES($1,$2,$3,$4,$5)", noteId, username, read, write, owner)
		if err != nil {
			log.Fatal(err)
		}
	}

}
func addNewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("addaccount.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()

		addUser(r.Form["username"][0], r.Form["password"][0])
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)

		//fmt.Println("password:", r.Form["password"])

		//defer r.Body.Close()

	}
}
func searchNotes(w http.ResponseWriter, r *http.Request) {
	var username, _ = r.Cookie("username")
	var TheNote *sql.Rows
	var matched bool

	r.ParseForm()

	//fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("search.html")

		t.Execute(w, nil)
	} else {
		db, err := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")

		rows, _ := db.Query(`SELECT * FROM PermissionsTable WHERE username = $1`, username.Value)
		var (
			NoteId   int
			Username string
			Read     bool
			Write    bool
			Owner    bool
		)
		for rows.Next() {
			err = rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)
			if Read == true {

				TheNote, _ = db.Query(`SELECT note, noteid FROM NotesTable`)

				if err != nil {
					log.Fatal(err)
				}

			}
		}

		userInput := r.Form["textboxid"][0]
		matched = false
		fmt.Println(userInput)
		option := r.Form["selectid"][0]
		switch option {
		case "prefix":

		case "phoneNumber":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				err = TheNote.Scan(&note, &noteid)

				matched, err = regexp.MatchString("^"+userInput+"+", note)

				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
				}

			}
		case "email":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				err = TheNote.Scan(&note, &noteid)

				matched, err = regexp.MatchString(".+@"+userInput+"+\\..+$", note)

				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
				}

			}

		case "text":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				err = TheNote.Scan(&note, &noteid)

				match := regexp.MustCompile("meeting|minutes|agenda|action|attendees|apologies")

				matches := match.FindAllStringIndex(note, -1)

				fmt.Println(len(matches))
				if len(matches) >= 3 {
					fmt.Fprintf(w, "<p>"+note+"</p>")
				}

			}
			fmt.Println("Text")
		case "capitals":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				err = TheNote.Scan(&note, &noteid)

				matched, _ = regexp.MatchString("([A-Z]){3,}", note)
				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
				}

			}

		default:
			fmt.Println("nothing")
		}

	}
}

func addNewNote(w http.ResponseWriter, r *http.Request) {
	var username, err = r.Cookie("username")
	if err == nil {
		fmt.Println(username.Value)

	} else {
		fmt.Println(err)
	}

	r.ParseForm()

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
	db, _ := sql.Open("postgres", "user=postgres password=chur dbname=webAppDatabase sslmode=disable")
	var username, err = r.Cookie("username")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(username)

	r.ParseForm()
	fmt.Println("method:", r.Method) //get request method

	if r.Method == "GET" {
		t, _ := template.ParseFiles("notes.html")

		t.Execute(w, nil)

	}

	if r.Method == "POST" {
		fmt.Println(r.Form)
		if r.Form.Get("Delete Note") == "Delete Note" {
			idstring := r.Form["aid"][0]
			i, err := strconv.Atoi(idstring)
			if err != nil {
				log.Fatal(err)
			}
			deleteNote(username.Value, i, db)
			t, _ := template.ParseFiles("notes.html")

			t.Execute(w, nil)
		} else {

			n := r.Form["anote"][0]
			idstring := r.Form["aid"][0]
			i, err := strconv.Atoi(idstring)
			if err != nil {
				log.Fatal(err)
			}

			updateNote(n, i, db)
			t, _ := template.ParseFiles("notes.html")

			t.Execute(w, nil)
		}
	}

	rows, err := db.Query(`SELECT * FROM PermissionsTable WHERE username = $1`, username.Value)
	var (
		NoteId   int
		Username string
		Read     bool
		Write    bool
		Owner    bool
	)

	for rows.Next() {
		err = rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)
		if Read == true && Write == true && Owner == true {
			//fmt.Println("note read write owner")
			//fmt.Println(NoteId)
			TheNote, err := db.Query(`SELECT note, username, noteid FROM NotesTable WHERE noteid = $1`, NoteId)
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println(TheNote)
			var (
				note     string
				username string
				noteid   int
			)
			for TheNote.Next() {

				err = TheNote.Scan(&note, &username, &noteid)
				fmt.Fprintf(w, "<h1>noteowner "+username+"</h1>")
				idAsString := strconv.Itoa(noteid)

				fmt.Fprintf(w, "<form action=\"/notes\" method=\"post\">"+
					"<textarea name=\"anote\"  cols=\"40\" rows=\"5\">"+note+"</textarea>"+

					"<input name=\"aid\" type=\"hidden\"value="+idAsString+">"+
					"<input type=\"submit\" value=\"Update Note\">"+
					"<input type=\"submit\" name=\"Delete Note\" value=\"Delete Note\">"+
					"</form>")

			}

		} else if Read == true && Write == true {
			fmt.Println("can read and write but not owner")
		} else if Read == true {
			fmt.Println("can read note")
			fmt.Println(NoteId)
		} else {
			fmt.Println("can't read note")
			fmt.Println(NoteId)
		}

	}

	/*
		rows, err := db.Query("SELECT * FROM NotesTable ")
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

			//fmt.Println(Username)
			//fmt.Println(Read)

			//fmt.Fprintf(w, Note+"<br>")

		}
	*/
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

//sets up all the tables and columns in the database//
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
	_, err = db.Exec("CREATE TABLE NotesTable(noteId SERIAL, username varchar(50), note text)")
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
