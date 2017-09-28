package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

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
	//addUser(data.Username, data.Password) //adds new user to the database
	fmt.Println(data.Username)
	fmt.Println(data.Password)
	defer r.Body.Close()

}
func addNewNote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	decoder := json.NewDecoder(r.Body)
	data := struct {
		NoteID   int    `json:"noteid"` //neeeds to be automatically generated
		Username string `json:"username"`
		Note     string `json:"note"`
	}{}
	err := decoder.Decode(&data)

	if err != nil {
		panic(err)
	}
	//addNote(data.NoteID, data.Username, data.Note) //adds new note to the database //todo add permissions also
	defer r.Body.Close()
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello Alex!")
}

func sayhelloPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}

	decoder := json.NewDecoder(r.Body)
	fmt.Println(decoder)
	data := struct {
		Message string `json:"msg"`
	}{}

	err := decoder.Decode(&data)

	if err != nil {
		panic(err)
	}
	fmt.Println(data.Message)
	defer r.Body.Close()
	fmt.Fprintf(w, string(data.Message))
}

func main() {
	http.HandleFunc("/adduser", addNewUser)
	http.HandleFunc("/addnote", addNewNote)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

/*
func (srv *Server) Serve(l net.Listener) error {
defer l.Close()
var tempDelay time.Duration
for{
	rw, e:= l.Accept()
	if e != nil{
		if ne, ok :=e.(net.Error); ok && ne.Temporary(){
			if tempDelay == 0 {
				tempDelay = 5 * time.Millisecond
			}else{
				tempDelay *= 2
			}
			if max := 1 * time.Second; tempDelay > max {
				tempDelay = max
			}
			log.Printf("http: Accept error: %v; retrying in %v", e, tempDelay)
			time.Sleep(tempDelay)
			continue
		}
		return e
	}
	tempDelay = 0
	c, err := srv.newConn(rw)
	if err != nil {
		continue
	}
	go c.serve()
}
}
*/
