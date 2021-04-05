/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

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




	// Get past orders
	var past_order_string string
	var past_orders_str_arrs []string
	err = db.QueryRow("SELECT past_orders FROM users WHERE phone=$1", pn).Scan(&past_order_string)
	if err != nil {
		return nil, err
	}
	past_orders_str_arrs = DBPGStringArrayToStringSlice(past_order_string)
	log.Printf("[ORDERHISTORY] %v", past_orders_str_arrs)

	orders := []*Order{}


	for _, order_id_str := range past_orders_str_arrs {
		order_id, _ := strconv.ParseInt(order_id_str, 10, 64)


		// get rest id
		var rest_id string
		err = db.QueryRow("SELECT rest_id FROM orders where order_id=$1", order_id).Scan(&rest_id)

		// get rest name
		var rest_name string

		err = db.QueryRow("SELECT rest_nameFROM restaurants where rest_id=$1", order_id).Scan(&rest_name)
		orderitems := []*OrderItem{}


		//get total tips first

		var total_order_tip float32
		total_order_tip = 0.0
		rows, err := db.Query("SELECT tx_id, tip from tx where paid_by = $1", pn)
		if err != nil {
			// handle this error better than this
			return &GetUserOrderHistoryResponse{}, err
		}
		defer rows.Close()
		for rows.Next() {
			var tx_id int64
			var tx_tip float32
			err = rows.Scan(&tx_id, &tx_tip)
			if err != nil {
				return &GetUserOrderHistoryResponse{}, err
			}
			total_order_tip = total_order_tip + tx_tip

		}


		rows, err = db.Query("SELECT item_name, item_type, item_cost, item_id, paid_for, total_splits, paid_by, order_id FROM orderitems where $1 = ANY(paid_by) AND order_id = $2", pn, order_id)
		if err != nil {
			// handle this error better than this
			return &GetUserOrderHistoryResponse{}, err
		}
		defer rows.Close()

		//running total of cost
		var order_total float32 = 0.0

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



			if total_splits > 0 {
				order_total = order_total + (item_cost/float32(total_splits))
			} else {
				order_total = order_total + item_cost

			}

			orderitems = append(orderitems, &OrderItem{Name: item_name, Type: item_type, Cost: item_cost, Id: item_id, PaidFor: paid_for, TotalSplits: total_splits, PaidBy: DBPGStringArrayToStringSlice(paid_by), OrderId: order_id})
		}
		err = rows.Err()
		if err != nil {
			return &GetUserOrderHistoryResponse{}, err
		}

		var tr float32

		//@TODO: actual tax stuff
		tr =  .08
		// so it doesnt  show up in order history if no items
		if len(orderitems) < 1 {
			tr = 0.0
		}


		orders = append(orders, &Order{RestName:rest_name, OrderId: order_id, Orders: orderitems, TaxRate: tr, TaxAmount: tr*order_total, Tip: total_order_tip})
	}



	return &GetUserOrderHistoryResponse{Orders: orders}, nil
}
// returns tip for phone number and item_id
func DBGetTip(phone string, id int64) (float32, error) {
	var tip float32
	log.Printf("Getting tip for %v %v", phone, id)
	err := db.QueryRow(`SELECT tip FROM tx WHERE paid_by=$1 AND item_id=$2`,
		phone, id).Scan(&tip)
	if err != nil {
		return 0, err
	}

	return tip, nil
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
	var menu_url string
	var rest_id int
	var table_id int
	var order_id int64
	var LEYE_id int64
	fmt.Println("Prepping Order")
	err := db.QueryRow(`SELECT rest_id, table_id, order_id FROM tokens WHERE token_code=$1`,
		in.TableToken).Scan(&rest_id, &table_id, &order_id)
	if err != nil {
		// handle this error better than this
		fmt.Println(in)
		log.Println(err)
		return &OrderInitiateResponse{}, status.Errorf(codes.NotFound, "Token Code Probably doesn't exist")
	}

	// Check if LEYE id is null

	//@TODO: Check if fields exist more elgantly
	var LEYE_id_null_count int
	err = db.QueryRow(`SELECT COUNT(*) FROM restaurants WHERE rest_id=$1 and LEYE_id is NULL`, rest_id).Scan(&LEYE_id_null_count);
	if err != nil {
		log.Fatal(err)
	}
	if LEYE_id_null_count > 0 {
		err = db.QueryRow(`SELECT rest_name, menu_url FROM restaurants WHERE rest_id=$1`,
			rest_id).Scan(&rest_name, &menu_url)
		if err != nil {
			// handle this error better than this
			fmt.Println(in)

			return &OrderInitiateResponse{}, status.Errorf(codes.NotFound, "Couldn't get restaurant info.")
		}
	} else {
		err = db.QueryRow(`SELECT rest_name, LEYE_id, menu_url FROM restaurants WHERE rest_id=$1`,
			rest_id).Scan(&rest_name, &LEYE_id, &menu_url)
		if err != nil {
			// handle this error better than this
			fmt.Println(in)

			return &OrderInitiateResponse{}, status.Errorf(codes.NotFound, "Couldn't get restaurant info")
		}

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

	rows, err := db.Query("SELECT item_name, item_type, item_cost, item_id, paid_for, total_splits, paid_by FROM orderitems where order_id=$1", order_id)
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
		var paid_by_str string
		err = rows.Scan(&item_name, &item_type, &item_cost, &item_id, &paid_for, &total_splits, &paid_by_str)
		if err != nil {
			return &OrderInitiateResponse{}, err
		}



		// Grab Initials
		paid_by_phone := DBPGStringArrayToStringSlice(paid_by_str)
		paid_by_name := make([]string,0)

		for _, phone := range paid_by_phone {
			if phone != "" {
				fname, lname, err := DBPhoneToFirstLastName(phone)
				if err != nil {
					return &OrderInitiateResponse{}, err
				}
				paid_by_name = append(paid_by_name, fname+" "+lname)
			}
		}

		// Hack so "" is not sent as phone number
		paid_by_phone_real := make([]string, 0)
		for _, phone := range paid_by_phone {
			if phone != "" {
				paid_by_phone_real =  append(paid_by_phone_real, phone)
			}
		}


		orderitems = append(orderitems, &OrderItem{Name: item_name, Type: item_type, Cost: item_cost, Id: item_id, PaidFor: paid_for, TotalSplits: total_splits, PaidBy: paid_by_phone_real, PaidByName: paid_by_name})

	}
	err = rows.Err()
	if err != nil {
		return &OrderInitiateResponse{}, err
	}


	// hack for floating point shit
	var tr float32
	tr =  .08
	ord := &Order{RestName: rest_name, OrderId: order_id, Orders: orderitems, TaxRate: tr, LeyeId: LEYE_id, MenuUrl: menu_url}



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

	// iterate through item pays

	//@TODO: Maybe will need to have some kinda "not_accepted" array returned to client instead of erroring out early




	for _, itempay := range in.ItemPay {

		var paid_for bool
		err = db.QueryRow("SELECT paid_for FROM orderitems WHERE item_id=$1", itempay.Id).Scan(&paid_for)

		if paid_for {
			return &OrderPayResponse{}, status.Errorf(codes.AlreadyExists, "Item already paid for: %v", itempay.Id)
		}

		var total_splits int64
		var current_cost float64
		var order_id int64
		err = db.QueryRow(`SELECT total_splits, item_cost, order_id FROM orderitems WHERE item_id=$1`,
			itempay.Id).Scan(&total_splits, &current_cost, &order_id)
		if err != nil {
			// handle this error better than this
			return &OrderPayResponse{}, err
		}

		// temp paid for variable
		pf := true

		if itempay.Split {
			total_splits = total_splits + 1
			pf = false
		}

		stmt, err := db.Prepare(`UPDATE orderitems SET paid_for=$1, total_splits=$2, paid_by= array_append(paid_by,$3) WHERE item_id=$4`)
		if err != nil {
			return &OrderPayResponse{}, err
		}

		_, err = stmt.Exec(pf, total_splits, pn, itempay.Id)
		if err != nil {
			return &OrderPayResponse{}, err
		}


		//check if user already has order in past order. Likely cuz they rescanned and are trying to purchase a new item
		// Get past orders
		var past_order_string string
		var past_orders_str_arrs []string
		err = db.QueryRow("SELECT past_orders FROM users WHERE phone=$1", pn).Scan(&past_order_string)
		if err != nil {
			return nil, err
		}
		past_orders_str_arrs = DBPGStringArrayToStringSlice(past_order_string)


		order_id_str := strconv.FormatInt(order_id, 10)

		if !DBStringInSlice(order_id_str, past_orders_str_arrs) {

			stmt, err = db.Prepare(`UPDATE users SET past_orders = array_append(past_orders,$1) WHERE phone=$2`)
			if err != nil {
				return &OrderPayResponse{}, err
			}

			_, err = stmt.Exec(strconv.FormatInt(order_id, 10), pn)
			if err != nil {
				return &OrderPayResponse{}, err
			}
		}


	}


	// get order id for tx table

	var order_id int64
	err = db.QueryRow(`SELECT order_id FROM orderitems WHERE item_id=$1`,
		in.ItemPay[0].Id).Scan(&order_id)
	if err != nil {
		// handle this error better than this
		return &OrderPayResponse{}, err
	}

	// ADD TO TRANSACTION
	log.Printf("Adding tip: %v %v %v",order_id, pn, in.Tip)
	stmt, err := db.Prepare(`INSERT INTO tx(order_id, paid_by, tip) VALUES($1, $2, $3)`)
	if err != nil {
		return &OrderPayResponse{}, err
	}
	_, err = stmt.Exec(order_id, pn, in.Tip)
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

func DBPhoneToFirstLastName(phone string) (string, string, error) {
	var DBFname, DBLname string
	log.Printf("Getting name for %v", phone)
	err := db.QueryRow(`SELECT fname, lname FROM users WHERE phone=$1`, phone).Scan(&DBFname, &DBLname);
	if err != nil {
		return "", "", err
	}
	return DBFname, DBLname, nil
}

func DBStringInSlice(a string, list []string) bool {
	for _, b := range list {

		//bint, _ := strconv.ParseInt(b, 10, 64)

		if b == a {
			return true
		}
	}
	return false
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

func DBSelectionLock_Update(in *SelectionRequest, val bool) error {
	stmt, err := db.Prepare(`UPDATE orderitems SET selected_by_lock = $1 WHERE item_id=$2`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(val, in.Id)
	if err != nil {
		return err
	}
	return nil
}
func DBClearSelects(in *SelectionCurrentUsersRequest) error {
	phone, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return err
	}
	stmt, err := db.Prepare(`UPDATE orderitems SET selected_by ='', selected_by_lock=false WHERE selected_by=$1`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(phone)
	if err != nil {
		return err
	}
	return nil
}

func DBSelectionClick(in *SelectionRequest) error {
	phone, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return err
	}

	//


	// Check if already paid fro
	var paid_for bool
	err  =  db.QueryRow("SELECT paid_for FROM orderitems WHERE item_id=$1", in.Id).Scan(&paid_for)
	if paid_for {
		return status.Errorf(codes.AlreadyExists, "Item already paid for: %v", in.Id)
	}

	// handle splits differently
	// @TODO: Handle splits differently
	//if !in.IsSplit {
		// is selected
		if in.IsSelected {
			// Check if already selected_by
			var selected_by string
			var selected_by_lock bool
			err := db.QueryRow(`SELECT selected_by, selected_by_lock FROM orderitems WHERE item_id=$1`, in.Id).Scan(&selected_by, &selected_by_lock)
			if err != nil {
				return err
			}


			// attempted at selection race conditions or smth
			if selected_by_lock  {
				log.Printf("[SELECT] Select Already Exists %v")
				return status.Errorf(codes.AlreadyExists, fmt.Sprintf("Already Selected %v", in.Id))
			}

			// if not splitting and already selected
			if !in.IsSplit && selected_by != "" {
				log.Printf("[SELECT] Select Already Exists %v")
				return status.Errorf(codes.AlreadyExists, fmt.Sprintf("Already Selected %v", in.Id))
			}


			//lock,
			log.Printf("[SELECT] locking")
			err = DBSelectionLock_Update(in, true);
			if err != nil {
				return err
			}

			if !in.IsSplit {

				stmt, err := db.Prepare(`UPDATE orderitems SET selected_by = $1 WHERE item_id=$2`)
				if err != nil {
					return err
				}

				_, err = stmt.Exec(phone, in.Id)
				if err != nil {
					return err
				}
			} else {
				stmt, err := db.Prepare(`UPDATE orderitems SET split_by=array_append(split_by, $1) WHERE item_id=$2`)
				if err != nil {
					return err
				}

				_, err = stmt.Exec(phone, in.Id)
				if err != nil {
					return err
				}
			}
		} else {
			//is unselected
			stmt, err := db.Prepare(`UPDATE orderitems SET selected_by = $1, selected_by_lock=false WHERE item_id=$2`)
			if err != nil {
				return err
			}

			_, err = stmt.Exec("", in.Id)
			if err != nil {
				return err
			}

			//is unselected
			stmt, err = db.Prepare(`UPDATE orderitems SET split_by = array_remove(split_by,$1), selected_by_lock=false WHERE item_id=$2`)
			if err != nil {
				return err
			}

			_, err = stmt.Exec(phone, in.Id)
			if err != nil {
				return err
			}


		}
	//}

	return nil
}

func DBGetSelects(in *SelectionCurrentUsersRequest) (*SelContArray, error) {
	ret := SelContArray{}



	return &ret, nil
	return &SelContArray{}, nil
}

func DBIsPhoneInDB(phone string) bool {
	var pn string
	err  :=  db.QueryRow("SELECT phone FROM users WHERE phone=$1", phone).Scan(&pn)
	if err != nil {
		log.Printf("No user with phone %v", phone)
		return false
	}
	return true

}

//@TODO: field exists