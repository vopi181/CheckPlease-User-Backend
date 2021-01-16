package CPUser

import (
	"context"
	"fmt"
	"github.com/vopi181/CheckPlease-User-Backend/CPUser/phone_auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)


type Server struct {
}



// Proto Verify
func VerifyCreateUserRequest(in *CreateUserRequest) error {
	//@TODO: add more userinput validate

	if(in.Phone == "" || in.Fname == "" || in.Lname == "") {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid User Data", in.String()));

	}





	return nil;
}

func (s *Server) CreateUser(ctx context.Context, in *CreateUserRequest) (*AuthTokenResponse, error){

	err := VerifyCreateUserRequest(in);

	if(err != nil) {
		return nil, err;
	}


	err = DBCreateUser(in);
	if(err != nil) {
		return nil, err;
	}
	log.Printf("Created User Account for %v", in.Phone);



	// Get the sms verification code and send the text to confirm later
	SMSVerificationCode, err := phone_auth.SendTextVerification(in.Phone);
	if(err != nil) {
		return nil, err;
	}

	err = DBUpdateTextVerificationToken(in.Phone, SMSVerificationCode)
	if(err != nil) {
		return nil, err;
	}

	return &AuthTokenResponse{AuthToken: ""}, nil
}


func (s *Server) SMSVerification(ctx context.Context, in *VerifySMSRequest) (*AuthTokenResponse, error) {
	if(in.Phone == "") {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid Phone Number: %v", in.String()));
	}

	DBSMSVerificationToken, err := DBGetTextVerificationToken(in);
	if(err != nil) {
		return nil, err
	}

	if DBSMSVerificationToken == in.SMSVerificationToken {
		tok := GenerateAuthToken(in.Phone)
		err = DBUpdateAuthToken(tok, in.Phone)
		if err != nil {
			return nil, err
		}
		return &AuthTokenResponse{AuthToken: tok}, nil
	} else {
		return nil, status.Errorf(codes.PermissionDenied, fmt.Sprintf("Invalid SMS Token: %v", in.SMSVerificationToken))
	}

}

//@TODO: update for sms
func (s *Server) ReAuthUser(ctx context.Context, in *ReAuthUserRequest) (*AuthTokenResponse, error) {

	SMSVerificationCode, err := phone_auth.SendTextVerification(in.Phone);
	if(err != nil) {
		return nil, err;
	}

	err = DBUpdateTextVerificationToken(in.Phone, SMSVerificationCode)
	if(err != nil) {
		return nil, err;
	}

	return &AuthTokenResponse{AuthToken: ""}, nil
}


func (s *Server) PaymentAddCard(ctx context.Context, in *PaymentAddCardRequest) (*PaymentAddCardResponse, error) {
	err := DBPaymentAddCard(in)
	if err != nil {
		return nil, err
	}

	return &PaymentAddCardResponse{}, nil
}

func (s *Server) GetUserInfo(ctx context.Context, in *AuthTokenRequest) (*UserInfoResponse, error) {
	UIR, err := DBGetUserInfo(in)
	if err != nil {
		return nil, err
	}

	return UIR, nil
}

func (s* Server) GetUserOrderHistory(ctx context.Context, in *AuthTokenRequest) (*GetUserOrderHistoryResponse, error) {
	UOHR, err := DBGetUserOrderHistory(in)
	if err != nil {
		return nil, err
	}

	return UOHR, nil
}


// ORDERS
func (s* Server) OrderInitiation(ctx context.Context, in *OrderInitiateRequest) (*OrderInitiateResponse, error) {
	OIR, err := DBPrepOrder(in)
	if err != nil {
		return nil, err
	}
	return OIR, nil
}

func (s* Server) OrderPay(ctx context.Context, in *OrderPayRequest) (*OrderPayResponse, error) {
	OPR, err := DBPayItem(in)
	if err != nil {
		return nil, err
	}
	return OPR, nil
}




// Simple Ping -> Pong  sanity check
func (s *Server) Ping(ctx context.Context, in *emptypb.Empty) (*PongResponse, error) {
	log.Printf("Ping Ponged");
	return &PongResponse{PongMessage: "Pong!"}, nil
}