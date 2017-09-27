package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

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

func sayhelloConnor(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello Connor!")
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
	fmt.Println(data)
	defer r.Body.Close()
	fmt.Fprintf(w, string(data.Message))
}

func main() {
	http.HandleFunc("/", sayhelloName)
	http.HandleFunc("/c", sayhelloConnor)
	http.HandleFunc("/newnote", sayhelloPost)
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
