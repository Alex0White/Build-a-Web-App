package main

func main() {

}

type User struct {
	username float64
	password float64
}
type Note struct {
	username string
	noteId   float64
	note     string
}
type Permissions struct {
	noteId   int
	username int
	read     bool
	write    bool
}
