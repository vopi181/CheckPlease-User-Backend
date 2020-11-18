package CPUser

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)
import _ "github.com/jackc/pgx/v4/stdlib"

var db *sql.DB

const (
	BCRYPT_DEFAULT_COST = 10
)


func DBCreateDBConn() error {
	//@TODO: Change to env db string
	TempDB, err := sql.Open("pgx", "host=localhost port=5432 user=postgres password=postpass  dbname=cpuser")
	if err != nil {
		log.Fatal(err)
	}
	//defer db.Close()
	db = TempDB;
	return nil;
}




// Creates user in database
// Encrypts with brcrypt
// Verify no existing user
func DBCreateUser(in *CreateUserRequest) error {

	// Make sure existing user is in DB
	//@TODO: optimize sql checking if user exist
	var existCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE username=$1`, in.Username).Scan(&existCount);
	if err != nil {
		log.Fatal(err)
	}
	if existCount > 0 {
		return status.Errorf(codes.AlreadyExists, fmt.Sprintf("User Already Exists"))
	}
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE email=$1`, in.Email).Scan(&existCount);
	if err != nil {
		log.Fatal(err)
	}
	if existCount > 0 {
		return status.Errorf(codes.AlreadyExists, fmt.Sprintf("User Already Exists"))
	}




	//@TODO: add  actual db error handling
	stmt, err := db.Prepare("INSERT INTO users(username, password, email) VALUES($1,$2,$3)")
	if err != nil {
		log.Fatal(err)
	}

	HashedPasswordBinary, err := bcrypt.GenerateFromPassword([]byte(in.Password), BCRYPT_DEFAULT_COST)

	_, err = stmt.Exec(in.Username,string(HashedPasswordBinary), in.Email)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}


func DBValidateAuth(in *ReAuthUserRequest) error {

	//@TODO: remove this not needed
	//// Turn submitted user password into brypt byte pass
	//UserPassBinary, err := bcrypt.GenerateFromPassword([]byte(in.Password), BCRYPT_DEFAULT_COST)
	//if(err != nil) {
	//	log.Fatal(err)
	//}



	// Get stored DB bcrypt pass in string format
	var DBPassString string
	err := db.QueryRow(`SELECT password FROM users WHERE username=$1`, in.Username).Scan(&DBPassString)
	if err != nil {
		return err
	}
	DBPassBinary := []byte(DBPassString)

	err = bcrypt.CompareHashAndPassword(DBPassBinary, []byte(in.Password))
	if err != nil {
		return status.Errorf(codes.PermissionDenied, fmt.Sprintf("Invalid Password"))
	}

	// password is correct
	return nil
}

func DBUpdateAuthToken(tok string, username string) error {
	stmt, err := db.Prepare(`UPDATE users SET auth_token=$1 WHERE username=$2`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(tok,username)
	if err != nil {
		return err
	}

	return nil

}
