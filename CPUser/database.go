package CPUser

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

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


//@TODO: Get user info
func DBGetUserInfo(in *AuthTokenRequest) (*UserInfoResponse, error) {
	var fname string;
	var lname string;
	var pn string;

	//ccs
	//var ccfname string
	//var cclname string
	//var num string
	//var cvv int32
	//var exp string


	err := db.QueryRow(`SELECT fname, lname, phone FROM users WHERE auth_token=$1`,
		in.Token).Scan(&fname, &lname, &pn)
	if err != nil {
		// handle this error better than this
		return &UserInfoResponse{}, err
	}
	//err = db.QueryRow(`SELECT fname, lname, cvv, exp FROM users WHERE num=$1`,
	//	num).Scan(&ccfname, &cclname, &cvv, &exp)
	//if err != nil {
	//	// handle this error better than this
	//	return &UserInfoResponse{}, err
	//}

	//pc := &PaymentCard{Fname: "", Lname: "", Num: "",
	//	Cvv: 0, Exp: "" }

	return &UserInfoResponse{Fname: fname, Lname: lname, Pn: pn}, nil


}


func DBGetUserOrderHistory(in *AuthTokenRequest) (*GetUserOrderHistoryResponse, error) {
	pn, err := DBAuthTokenToPhone(in.Token)
	if err != nil {
		return nil, err
	}

	orderitems := []*OrderItem{}

	rows, err := db.Query("SELECT item_name, item_type, item_cost, item_id, paid_for, total_splits, paid_by, order_id FROM orderitems where $1 = ANY(paid_by)", pn)
	if err != nil {
		// handle this error better than this
		return &GetUserOrderHistoryResponse{}, err
	}
	defer rows.Close()
	// Iterate through rows  in DB
	for rows.Next() {
		fmt.Println("Iterating order item rows")
		var item_name string
		var item_type string
		var item_cost float32
		var item_id int64
		var paid_for bool
		var total_splits int64
		var paid_by string
		var order_id int64
		err = rows.Scan(&item_name, &item_type, &item_cost, &item_id, &paid_for, &total_splits, &paid_by, &order_id)
		if err != nil {
			return &GetUserOrderHistoryResponse{}, err
		}

		orderitems = append(orderitems, &OrderItem{Name: item_name, Type: item_type, Cost: item_cost, Id: item_id, PaidFor: paid_for, TotalSplits: total_splits, PaidBy: DBPGStringArrayToStringSlice(paid_by), OrderId: order_id})

	}
	err = rows.Err()
	if err != nil {
		return &GetUserOrderHistoryResponse{}, err
	}

	return &GetUserOrderHistoryResponse{Orders: orderitems}, nil
}



// ###### PAYMENT ######
//@TODO: add hanlding of primary  cards
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



// ###### ORDERS ######
func DBPrepOrder(in *OrderInitiateRequest) (*OrderInitiateResponse, error) {
	var rest_name string
	var rest_id int
	var table_id int
	var order_id int64

	fmt.Println("Prepping Order")
	err := db.QueryRow(`SELECT rest_name, rest_id, table_id, order_id FROM tokens WHERE token_code=$1`,
		in.TableToken).Scan(&rest_name, &rest_id, &table_id, &order_id)
	if err != nil {
		// handle this error better than this
		fmt.Println(in)

		return &OrderInitiateResponse{}, err
	}

	stmt, err := db.Prepare(`UPDATE users SET current_order=$1 WHERE auth_token=$2`)
	if err != nil {
		return &OrderInitiateResponse{}, err
	}

	_, err = stmt.Exec(order_id,in.AuthRequest.Token)
	if err != nil {
		return &OrderInitiateResponse{}, err
	}

	orderitems := []*OrderItem{}

	rows, err := db.Query("SELECT item_name, item_type, item_cost, item_id, paid_for, total_splits FROM orderitems where order_id=$1", order_id)
	if err != nil {
		// handle this error better than this
		return &OrderInitiateResponse{}, err
	}
	defer rows.Close()
	// Iterate through rows  in DB
	for rows.Next() {
		fmt.Println("Iterating order item rows")
		var item_name string
		var item_type string
		var item_cost float32
		var item_id int64
		var paid_for bool
		var total_splits int64
		err = rows.Scan(&item_name, &item_type, &item_cost, &item_id, &paid_for, &total_splits)
		if err != nil {
			return &OrderInitiateResponse{}, err
		}

		orderitems = append(orderitems, &OrderItem{Name: item_name, Type: item_type, Cost: item_cost, Id: item_id, PaidFor: paid_for, TotalSplits: total_splits})

	}
	err = rows.Err()
	if err != nil {
		return &OrderInitiateResponse{}, err
	}
	ord := &Order{RestName: rest_name, OrderId: order_id, Orders: orderitems}

	return &OrderInitiateResponse{Order: ord}, nil

}
//@TODO: HANDLE PAYMENTS!!!!!!!!!!!!!!
func DBPayItem(in *OrderPayRequest) (*OrderPayResponse, error) {

	//get phone number for order history
	pn, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		// handle this error better than this
		return &OrderPayResponse{}, err
	}

	//handle payment


	// update db



	var total_splits int64
	var current_cost float64
	var order_id int64
	err = db.QueryRow(`SELECT total_splits, item_cost, order_id FROM orderitems WHERE item_id=$1`,
		in.ItemPay.Id).Scan(&total_splits, &current_cost, &order_id)
	if err != nil {
		// handle this error better than this
		return &OrderPayResponse{}, err
	}

	// temp paid for variable
	pf := true

	if in.ItemPay.Split {
		total_splits = total_splits + 1
		pf = false
	}


	stmt, err := db.Prepare(`UPDATE orderitems SET paid_for=$1, total_splits=$2, paid_by= array_append(paid_by,$3) WHERE item_id=$4`)
	if err != nil {
		return &OrderPayResponse{}, err
	}

	_, err = stmt.Exec(pf, total_splits, pn, in.ItemPay.Id)
	if err != nil {
		return &OrderPayResponse{}, err
	}

	stmt, err = db.Prepare(`UPDATE users SET past_orders = array_append(past_orders,$1) WHERE phone=$2`)
	if err != nil {
		return &OrderPayResponse{}, err
	}

	_, err = stmt.Exec(strconv.FormatInt(order_id, 10), pn)
	if err != nil {
		return &OrderPayResponse{}, err
	}

	return &OrderPayResponse{Accepted: true}, nil
}




// ##### SELECTION #####
//@TODO: Move to DB at some point
//func DBUpdateOrderItemSelectedBy(tok string, item_id int) error {
//
//}


// ###### HELPERS ######
func DBAuthTokenToPhone(tok string) (string, error) {
	var DBPhone string
	err := db.QueryRow(`SELECT phone FROM users WHERE auth_token=$1`, tok).Scan(&DBPhone)
	if err != nil {
		return "", err
	}

	return DBPhone, nil
}

func DBAuthTokenToFirstLastName(tok string) (string, string, error) {
	var DBFname, DBLname string
	err := db.QueryRow(`SELECT fname, lname FROM users WHERE auth_token=$1`, tok).Scan(&DBFname, &DBLname);
	if err != nil {
		return "", "", err
	}
	return DBFname, DBLname, nil
}


//func DBAuthTokenCompare(tok string) (bool, error) {
//	var DBAuthToken string
//	err := db.QueryRow(`SELECT auth_token FROM users WHERE auth_token=$1`, tok).Scan(&DBAuthToken)
//	if err != nil {
//		return "", err
//	}
//}



func DBPing() error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return nil
}


// HELPERS

func DBPGStringArrayToStringSlice(str string) []string {
	ret := strings.Split(str, ",")
	ret[0] = string(([]rune(ret[0]))[1:len(ret[0])])
	ret[len(ret)-1] = string(([]rune(ret[len(ret)-1]))[0:len(ret[len(ret)-1])-1])
	return ret;
}

func DBSelectionClick(in *SelectionRequest) error {
	phone, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return err
	}

	// handle splits differently
	// @TODO: Handle splits differently
	if !in.IsSplit {
		// is selected
		if in.IsSelected {
			// Check if already selected_by
			var selected_by string
			err := db.QueryRow(`SELECT selected_by FROM orderitems WHERE item_id=$1`, in.Id).Scan(&selected_by)
			if err != nil {
				return err
			}

			if selected_by != "" {
				return status.Errorf(codes.AlreadyExists, fmt.Sprintf("Already Selected %v", in.Id))
			}

			stmt, err := db.Prepare(`UPDATE orderitems SET selected_by = $1 WHERE item_id=$2`)
			if err != nil {
				return err
			}

			_, err = stmt.Exec(phone, in.Id)
			if err != nil {
				return err
			}

		} else {
			//is unselected
			stmt, err := db.Prepare(`UPDATE orderitems SET selected_by = $1 WHERE item_id=$2`)
			if err != nil {
				return err
			}

			_, err = stmt.Exec("", in.Id)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func DBGetSelects(in *SelectionCurrentUsersRequest) (*SelContArray, error) {
	ret := SelContArray{}



	return &ret, nil
	return &SelContArray{}, nil
}