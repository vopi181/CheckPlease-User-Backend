package CPUser

import (
	"database/sql"
	"fmt"
	//"golang.org/x/crypto/bcrypt"
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
// Verify no existing user
func DBCreateUser(in *CreateUserRequest) error {

	// Make sure existing user is in DB
	//@TODO: optimize sql checking if user exist
	var existCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE phone=$1`, in.Phone).Scan(&existCount);
	if err != nil {
		log.Fatal(err)
	}
	//if existCount > 0 {
	//	return status.Errorf(codes.AlreadyExists, fmt.Sprintf("User Already Exists"))
	//}
	//err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE email=$1`, in.Email).Scan(&existCount);
	//if err != nil {
	//	log.Fatal(err)
	//}
	if existCount > 0 {
		return status.Errorf(codes.AlreadyExists, fmt.Sprintf("User Already Exists"))
	}




	//@TODO: add  actual db error handling
	stmt, err := db.Prepare("INSERT INTO users(phone, fname, lname) VALUES($1,$2,$3)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(in.Phone,in.Fname, in.Lname)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

//@TODO: fix for phone
//func DBValidateAuth(in *ReAuthUserRequest) error {
//
//	//@TODO: remove this not needed
//	//// Turn submitted user password into brypt byte pass
//	//UserPassBinary, err := bcrypt.GenerateFromPassword([]byte(in.Password), BCRYPT_DEFAULT_COST)
//	//if(err != nil) {
//	//	log.Fatal(err)
//	//}
//
//
//
//	// Get stored DB bcrypt pass in string format
//	var DBPassString string
//	err := db.QueryRow(`SELECT auth_token FROM users WHERE phone=$1`, in.Phone).Scan(&DBPassString)
//	if err != nil {
//		return err
//	}
//	DBPassBinary := []byte(DBPassString)
//
//	err = bcrypt.CompareHashAndPassword(DBPassBinary, []byte(in.Password))
//	if err != nil {
//		return status.Errorf(codes.PermissionDenied, fmt.Sprintf("Invalid Password"))
//	}
//
//	// password is correct
//	return nil
//}

func DBUpdateAuthToken(tok string, phone string) error {
	stmt, err := db.Prepare(`UPDATE users SET auth_token=$1 WHERE phone=$2`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(tok,phone)
	if err != nil {
		return err
	}

	return nil

}

func DBUpdateTextVerificationToken(phone string, tok string) error {
	stmt, err := db.Prepare(`UPDATE users SET current_sms_verification_token=$1 WHERE phone=$2`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(tok,phone)
	if err != nil {
		return err
	}

	return nil
}

func DBGetTextVerificationToken(in *VerifySMSRequest) (string, error) {
	var DBSMSVerificationToken string
	err := db.QueryRow(`SELECT current_sms_verification_token FROM users WHERE phone=$1`, in.Phone).Scan(&DBSMSVerificationToken)
	if err != nil {
		return "", err
	}

	return DBSMSVerificationToken, nil;
}

// ###### PAYMENT ######
func DBPaymentAddCard(in *PaymentAddCardRequest) error {
	stmt, err := db.Prepare("INSERT INTO payinfo(fname, lname, num, cvv, exp, phone) VALUES($1,$2,$3,$4,$5,$6)")
	if err != nil {
		log.Fatal(err)
	}

	pn, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(in.Card.Fname,in.Card.Lname,in.Card.Num,in.Card.Cvv,in.Card.Exp,pn)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}


// ###### HELPERS ######
func DBAuthTokenToPhone(tok string) (string, error) {
	var DBPhone string
	err := db.QueryRow(`SELECT phone FROM users WHERE auth_token=$1`, tok).Scan(&DBPhone)
	if err != nil {
		return "", err
	}

	return DBPhone, nil
}

//func DBAuthTokenCompare(tok string) (bool, error) {
//	var DBAuthToken string
//	err := db.QueryRow(`SELECT auth_token FROM users WHERE auth_token=$1`, tok).Scan(&DBAuthToken)
//	if err != nil {
//		return "", err
//	}
//}