DROP TABLE IF EXISTS NotesTable;
DROP TABLE IF EXISTS PermissionsTable;
DROP TABLE IF EXISTS TempNoteIDTable;
DROP TABLE IF EXISTS UsersTable;


CREATE TABLE UsersTable(username varchar(50), password varchar(50),PRIMARY KEY(username));
CREATE TABLE NotesTable(noteId SERIAL, username varchar(50), note text);
CREATE TABLE PermissionsTable(noteId int, username varchar(50), read boolean, write boolean, owner boolean);
CREATE TABLE TempNoteIDTable(note text, noteId SERIAL, username varchar(50), FOREIGN KEY(username) REFERENCES UsersTable(username))

