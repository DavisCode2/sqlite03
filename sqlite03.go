/*
The package works on 2 tables on an SQLite database.

The names of the tables are:

  - Users
  - Userdata

The definitions of the tables are:

	CREATE TABLE Users (
	    ID INTEGER PRIMARY KEY,
	    Username TEXT
	);

	CREATE TABLE Userdata (
	    UserID INTEGER NOT NULL,
	    Name TEXT,
	    Surname TEXT,
	    Description TEXT
	);
*/
package sqlite03

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

/*
This global variable holds the SQLite3 database filepath

	Filename: Is the filepath to the database file
*/
var (
	Filename = ""
)

// The Userdata structure is for holding full user data
// from the Userdata table and the Username from the
// Users table
type UserData struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

// openConnection() is for opening the SQLite3 connection
// in order to be used by the other functions of the package.
func openConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", Filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// The function returns the User ID of the username
// -1 if the user does not exist

func exists(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := -1
	statement := fmt.Sprintf(`SELECT ID FROM Users WHERE Username = '%s'`, username)
	rows, err := db.Query(statement)
	if err != nil {
		fmt.Println("exists ", err)
		return -1
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("exists() Scan", err)
			return -1
		}
		userID = id
	}
	return userID
}

// AddUser function adds a new user to the sqlite database
// It returns the new User ID
// -1 if it was an error

func AddUser(d UserData) int {
	d.Username = strings.ToLower(d.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID != -1 {
		fmt.Println("User already exists: ", d.Username)
	}
	insertStatement := `INSERT INTO Users VALUES (NULL, ?)`
	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println("err")
		return -1
	}

	// Check that the insert statement is successful
	userID = exists(d.Username)
	if userID == -1 {
		return userID
	}

	insertStatement = `INSERT INTO UserData values (?, ?, ?, ?)`
	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec(): ", err)
		return -1
	}
	return userID
}

/*
DeleteUser deletes an existing user if the user exists.
It requires the User ID of the user to be deleted.

It returns error
*/
func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	statement := fmt.Sprintf(`SELECT Username FROM Users WHERE ID = %d`, id)
	rows, err := db.Query(statement)
	if err != nil {
		return err
	}
	defer rows.Close()

	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}

	// Check whether the username exists
	if exists(username) != id {
		return fmt.Errorf("user with ID %d does not exist", id)
	}

	// Delete from UserData
	deleteStatement := `DELETE FROM Userdata WHERE UserID = ?`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// Delete from Users
	deleteStatement = `DELETE FROM Users WHERE ID = ?`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	return nil
}

// ListUsers() lists all users in the database.
//
// Returns a slice of Userdata to the calling function.
func ListUsers() ([]UserData, error) {
	Data := []UserData{}
	db, err := openConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT ID, Username, Name, Surname, Description FROM Users, Userdata WHERE Users.ID = Userdata.UserID`)
	if err != nil {
		return Data, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var desc string

		err = rows.Scan(&id, &username, &name, &surname, &desc)
		temp := UserData{ID: id, Username: username, Name: name, Surname: surname, Description: desc}
		Data = append(Data, temp)
		if err != nil {
			return nil, err
		}
	}
	return Data, nil
}

/*
UpdateUser() is for updating an existing user
given a Userdata structure.
The user ID of the user to be updated is found
inside the function.
*/
func UpdateUser(d UserData) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exists(d.Username)
	if userID == -1 {
		return errors.New("user does not exist")
	}
	d.ID = userID
	updateStatement := `UPDATE Userdata set Name = ?, Surname = ?, Description = ? WHERE UserID = ?`
	_, err = db.Exec(updateStatement, d.Name, d.Surname, d.Description, d.ID)
	if err != nil {
		return err
	}
	return nil
}
