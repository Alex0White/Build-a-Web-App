package main

func main() {

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
