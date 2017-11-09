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
	http.HandleFunc("/", login)
	http.HandleFunc("/adduser", addNewUser)
	http.HandleFunc("/notes", viewNotes)
	http.HandleFunc("/createnote", addNewNote)
	http.HandleFunc("/search", searchNotes)
	http.HandleFunc("/changepermissions", changeNewPermissions)
	http.HandleFunc("/notepermissions", notePermissionsView)

	db = OpenDB()
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

var db *sql.DB

func OpenDB() *sql.DB {
	db, err := sql.Open("postgres", "user=postgres password=password dbname=webAppDatabase sslmode=disable port=5432 ")
	if err != nil {
		log.Fatal(err)
		fmt.Println("Someting went wrong while opening the database")
	}
	return db
}

func login(w http.ResponseWriter, r *http.Request) {
	var loggedin = false

	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()
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

func addNote(username string, note string, db *sql.DB) { //adds a new note to the database

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

func addUser(username string, password string, db *sql.DB) bool { //adds a new user to the database

	//check if username is already taken

	testuser := username

	user, err := db.Query("SELECT exists(SELECT username FROM UsersTable WHERE username = $1)", testuser)

	user.Next()
	var aUser bool
	user.Scan(&aUser)
	userExists := aUser

	fmt.Println(aUser)
	if err != nil {

		fmt.Println(err, "error thing for adduser == nil")
	}

	if userExists == false {
		db.Exec("INSERT INTO UsersTable(username, password) VALUES($1,$2)", username, password)
		return false
	} else {
		return true
	}

}

func changePermissions(noteId int, username string, read bool, write bool, owner bool, db *sql.DB) { //needs to be change permissions, if the user is already associated with the note
	updated := false

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

type user struct {
	name string
}
type permission struct {
	NoteId   int
	Username string
	Read     bool
	Write    bool
	Owner    bool
}
type notePermissionTemp struct {
	note        string
	noteId      int
	users       []user
	permissions []permission
}

func notePermissionsView(w http.ResponseWriter, r *http.Request) { //this adds content to the notepermissions page. This content shows the notes that the owners accout has created and gives them options to add permissions to their notes  
	var username, err = r.Cookie("username")
	if err != nil {
		log.Fatal(err)
	}

	r.ParseForm()

	if r.Method == "GET" {
		t, _ := template.ParseFiles("notePermissions.html")

		t.Execute(w, nil)

	}

	idstring := r.Form["aid"][0]
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("idstring: ", idstring)
	i, err := strconv.Atoi(idstring)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("this (note PermisssionsView)is note id of created note: ", i)
	currentNoteID = i
	t, _ := template.ParseFiles("notepermissions.html")
	t.Execute(w, nil)
	if r.Form.Get("Add Permissions") == "Add or Remove Permissions" {
		fmt.Println("add permissions button works")
		write := false
		read := false
		if r.Form["WritePriv"] != nil {
			if r.Form["WritePriv"][0] == "Write" {
				fmt.Println("WriteCheck box workked")
				write = true
			}
		}
		if r.Form["ReadPriv"] != nil {
			if r.Form["ReadPriv"][0] == "Read" {
				fmt.Println("read Checkbox workked")
				read = true
			}
		}
		if read == true || write == true {
			theUser := r.Form["addthisuser"][0]
			fmt.Println("this is the currnet note id: ", currentNoteID)

			changePermissions(currentNoteID, theUser, read, write, false, db)
		}
	}
	aStruct := notePermissions(currentNote, currentNoteID, db)
	fmt.Println(aStruct)

	fmt.Fprintf(w, "<h2>Note Owner: "+username.Value+"</h2>"+
		"<p>"+aStruct.note+"</p>"+"<form action=\"/notepermissions\" method=\"post\"><select name =\"addthisuser\">")
	for _, u := range aStruct.users {
		fmt.Fprintf(w, "<option value="+u.name+">"+u.name+"</option>")
	}

	fmt.Fprintf(w, "</select>"+
		"<input type=\"checkbox\" name=\"ReadPriv\" value=\"Read\">Read"+
		"<input type=\"checkbox\" name=\"WritePriv\" value=\"Write\">Write"+
		"<input name=\"aid\" type=\"hidden\"value="+idstring+">"+
		"<input type=\"submit\" name=\"Add Permissions\" value=\"Add or Remove Permissions\">"+
		"</form>")

}

func notePermissions(note string, noteId int, db *sql.DB) notePermissionTemp {

	thing := notePermissionTemp{}
	rows, _ := db.Query("SELECT username FROM userstable ")
	var username string

	for rows.Next() {

		rows.Scan(&username)
		fmt.Println(username)

		thing.users = append(thing.users, user{name: username})

	}
	thing.note = note
	thing.noteId = noteId
	permissionsRows, err := db.Query("SELECT * FROM permissionstable WHERE noteid = $1", noteId)
	if err != nil {
		log.Fatal(err)
	}
	var (
		nId      int
		uname    string
		reed     bool
		wriit    bool
		theOwner bool
	)
	fmt.Println(nId, uname, reed, wriit, theOwner)
	for permissionsRows.Next() {
		permissionsRows.Scan(&nId, &uname, &reed, &wriit, &theOwner)

		thing.permissions = append(thing.permissions, permission{NoteId: nId, Username: uname, Read: reed, Write: wriit, Owner: theOwner})

	}
	return thing
}
func addNewUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("addaccount.html")
		t.Execute(w, nil)
	} else {

		r.ParseForm()

		takenUser := addUser(r.Form["username"][0], r.Form["password"][0], db)
		if takenUser {
			t, _ := template.ParseFiles("addaccount.html")
			t.Execute(w, nil)
		} else {
			t, _ := template.ParseFiles("login.html")
			t.Execute(w, nil)
		}

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
		//db, err := sql.Open("postgres", "user=postgres password=password dbname=webAppDatabase sslmode=disable")
		tempNoteRows, _ := db.Query(`SELECT * FROM tempnoteidtable WHERE username = $1`, username.Value)
		var (
			tempNote string
			tempNoteid int
			tempUsername string

		)
		rows, _ := db.Query(`SELECT * FROM PermissionsTable WHERE username = $1`, username.Value)
		var (
			NoteId   int
			Username string
			Read     bool
			Write    bool
			Owner    bool
		)
		if r.Form.Get("research") == "Search Results"{
			for tempNoteRows.Next(){
			 tempNoteRows.Scan(&tempNote, &tempNoteid, &tempUsername)
			 if tempUsername == username.Value{
			 TheNote, _ = db.Query(`SELECT note, noteid FROM tempnoteidtable`)
			 fmt.Print("it lives")
			 }
			}
			db.Exec(`DELETE FROM tempnoteidtable WHERE username = $1`, username.Value)
		}else{
		for rows.Next() {
			rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)
			if Read == true  {

				TheNote, _ = db.Query(`SELECT note, noteid FROM NotesTable WHERE noteid = $1`, NoteId)
		
		

			}
		}
		db.Exec(`DELETE FROM tempnoteidtable WHERE username = $1`, username.Value)
	}
		userInput := r.Form["textboxid"][0]
		matched = false
		fmt.Println(userInput)
		option := r.Form["selectid"][0]


if TheNote == nil{
	fmt.Println("thenote is empty")
	t, err := template.ParseFiles("search.html")
	if err != nil{
		log.Fatal(err)
	}
	t.Execute(w, nil)
}else{
		switch option {
		case "prefix":
			t, err := template.ParseFiles("search.html")
			if err != nil{
				log.Fatal(err)
			}
			t.Execute(w, nil)
				for TheNote.Next() { //
					var (
						note   string
						noteid int
					)
					TheNote.Scan(&note, &noteid)

					matched, _ = regexp.MatchString("\\. "+userInput+"|^"+userInput, note)

					if matched {
						fmt.Fprintf(w, "<p>"+note+"</p>")
						db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)

					}

				}
			

			

		case "suffix":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				TheNote.Scan(&note, &noteid)

				matched, _ = regexp.MatchString(userInput+"\\.", note)

				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
					db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)
				
				}

			}

		case "phoneNumber":
			t, _ := template.ParseFiles("search.html")
			t.Execute(w, nil)
			for TheNote.Next() {
				var (
					note   string
					noteid int
				)
				TheNote.Scan(&note, &noteid)

				matched, _ = regexp.MatchString("\\D"+userInput+"\\d", note)

				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
					db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)
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
				TheNote.Scan(&note, &noteid)

				matched, _ = regexp.MatchString("\\w+@"+userInput+"+.*\\.\\w", note)

				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
					db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)
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
				TheNote.Scan(&note, &noteid)

				match := regexp.MustCompile("meeting|minutes|agenda|action|attendees|apologies")

				matches := match.FindAllStringIndex(note, -1)

				fmt.Println(len(matches))
				if len(matches) >= 3 {
					fmt.Fprintf(w, "<p>"+note+"</p>")
					db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)
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
				TheNote.Scan(&note, &noteid)

				matched, _ = regexp.MatchString("([A-Z]){3,}", note)
				if matched {

					fmt.Fprintf(w, "<p>"+note+"</p>")
					db.Exec("INSERT INTO TempNoteIDTable(note, noteid, username) VALUES($1,$2,$3)", note, noteid, username.Value)
				}

			}

		default:
			fmt.Println("nothing")
		}
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
		addNote(username.Value, r.Form["note"][0], db)

		t, _ := template.ParseFiles("createnote.html")
		t.Execute(w, nil)

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
	changePermissions(data.NoteId, data.Username, data.Read, data.Write, data.Owner, db)

	defer r.Body.Close()

}

func viewUsers(db *sql.DB) {
	//db, _ := sql.Open("postgres", "user=postgres password=password dbname=webAppDatabase sslmode=disable port=5432")
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

var currentNoteID int
var currentNote string

func viewNotes(w http.ResponseWriter, r *http.Request) { // adds content to the notes.html page. Shows notes user created and lets them update delete and change permissions. 

	var username, err = r.Cookie("username")
	if err != nil {
		log.Fatal(err)
	}

	r.ParseForm()
	fmt.Println("method:", r.Method) //get request method

	if r.Method == "GET" {
		t, _ := template.ParseFiles("notes.html")

		t.Execute(w, nil)

	}

	if r.Method == "POST" {
		idstring := r.Form["aid"][0]
		i, err := strconv.Atoi(idstring)
		fmt.Println("this is note id of created note: ", i)
		currentNoteID = i
		n := r.Form["anote"][0]
		currentNote = n
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(r.Form)
		if r.Form.Get("Delete Note") == "Delete Note" {

			deleteNote(username.Value, i, db)
			t, _ := template.ParseFiles("notes.html")

			t.Execute(w, nil)

		} else {
			updateNote(n, i, db)
			
			t, _ := template.ParseFiles("notes.html")

			t.Execute(w, nil)
		}
	}
	if r.Form.Get("Edit Permissions") != "Edit Permissions" {  

		rows, _ := db.Query(`SELECT * FROM PermissionsTable WHERE username = $1`, username.Value) // finds all the permissions the user is assoiated with and adds them to the rows variable
		var (
			NoteId   int
			Username string
			Read     bool
			Write    bool
			Owner    bool
		)

		for rows.Next() { // goes through each permission in rows and gets the notes in the notestable associated with those permissions. Then displays the notes according to what permissions they had.    
			err = rows.Scan(&NoteId, &Username, &Read, &Write, &Owner)
			if Read == true && Write == true && Owner == true { // for notes that are created by the user

				TheNote, err := db.Query(`SELECT note, username, noteid FROM NotesTable WHERE noteid = $1`, NoteId) //gets the note associated with the id in permissions
				if err != nil {
					log.Fatal(err)
				}

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
						"<textarea name=\"anote\"  cols=\"40\" rows=\"5\">"+note+"</textarea>"+"<br>"+
						"<input name=\"aid\" type=\"hidden\"value="+idAsString+">"+
						"<input type=\"submit\" value=\"Update Note\">"+
						"<input type=\"submit\" name=\"Delete Note\" value=\"Delete Note\"></form>")

					fmt.Fprintf(w, "<form action=\"/notepermissions\" method=\"post\"><input type=\"submit\" name=\"Edit Permissions\" value=\"Edit Permissions\">"+
						"<input name=\"aid\" type=\"hidden\"value="+idAsString+"></form>")

				}

			} else if Read == true && Write == true { // for note that the user has been given read and write access by another user
				fmt.Println("can read and write but not owner")
				TheNote, err := db.Query(`SELECT note, username, noteid FROM NotesTable WHERE noteid = $1`, NoteId)
				if err != nil {
					log.Fatal(err)
				}

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
						"<textarea name=\"anote\"  cols=\"40\" rows=\"5\">"+note+"</textarea>"+"<br>"+
						"<input name=\"aid\" type=\"hidden\"value="+idAsString+">"+
						"<input type=\"submit\" value=\"Update Note\">"+
						"</form>")

				}

			} else if Read == true { //for notes that the user has been given read access by another user
				fmt.Println("can read note")
				TheNote, err := db.Query(`SELECT note, username, noteid FROM NotesTable WHERE noteid = $1`, NoteId)
				if err != nil {
					log.Fatal(err)
				}

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
						"<textarea name=\"anote\"  cols=\"40\" rows=\"5\"disabled>"+note+"</textarea>"+"<br>"+
						"<input name=\"aid\" type=\"hidden\"value="+idAsString+">"+
						"</form>")

				}
			} else {
				fmt.Println("can't read note")
				fmt.Println(NoteId)
			}

		}
	}

}
func viewPermissions() {

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
	db, err := sql.Open("postgres", "user=postgres password=password dbname=webAppDatabase sslmode=disable port=5432")
	err = db.Ping()
	//TempNoteIDTable need to Drop this before userstable as it contains username as foreign key 
	_, err = db.Exec("DROP TABLE IF EXISTS TempNoteIDTable")
	if err != nil {
		log.Fatal(err)
	}
	//userstable

	_, err = db.Exec("DROP TABLE IF EXISTS UsersTable")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("CREATE TABLE UsersTable(username varchar(50), password varchar(50), PRIMARY KEY (username))")
	if err != nil {
		log.Fatal(err)
	}

	//notestable
	_, err = db.Exec("DROP TABLE IF EXISTS NotesTable")
	if err != nil {
		log.Fatal(err)
	}
	//TempNoteIDTable
	_, err = db.Exec("CREATE TABLE TempNoteIDTable(note text, noteId SERIAL, username varchar(50), FOREIGN KEY (username) REFERENCES UsersTable (username))")
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
