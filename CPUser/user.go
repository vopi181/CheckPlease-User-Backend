package CPUser

import (
	"context"
	"fmt"
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

	if(in.Username == "" || in.Email == "" || in.Password == "") {
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
	log.Printf("Created User Login for %v", in.Username);

	tok := GenerateAuthToken(in.Username)
	err = DBUpdateAuthToken(tok, in.Username)
	if err != nil {
		return nil, err
	}

	return &AuthTokenResponse{AuthToken: tok}, nil
}


func (s *Server) ReAuthUser(ctx context.Context, in *ReAuthUserRequest) (*AuthTokenResponse, error) {

	err := DBValidateAuth(in)
	if err != nil {
		return &AuthTokenResponse{AuthToken: ""}, err
	}
	tok := GenerateAuthToken(in.Username)
	err = DBUpdateAuthToken(tok, in.Username)
	if err != nil {
		return nil, err
	}

	return &AuthTokenResponse{AuthToken: tok}, nil
}



// Simple Ping -> Pong  sanity check
func (s *Server) Ping(ctx context.Context, in *emptypb.Empty) (*PongResponse, error) {
	log.Printf("Ping Ponged");
	return &PongResponse{PongMessage: "Pong!"}, nil
}