package CPUser

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vopi181/CheckPlease-User-Backend/CPUser/phone_auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)


type SelectionNotificationChan struct {
	tokenCode    string;
	//@TODO: Cache selects for late scanning user. Should move to DB.
	selectCache []SelectionContainer;
	chanMap map[string](chan SelectionContainer)
}

type ItemPayNotificationChan struct {
	tokenCode string
	tokenPays chan ItemPayNotification
}

type Server struct {
	selects []SelectionNotificationChan;
	pays []ItemPayNotificationChan;
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

	is_chan := false;

	for _, chans := range s.selects {
		if chans.tokenCode == in.TableToken {
			is_chan = true
			break
		}
	}

	if !is_chan {
		s.selects = append(s.selects, SelectionNotificationChan{tokenCode: in.TableToken,  selectCache: []SelectionContainer{}, chanMap: make(map[string](chan SelectionContainer))})
	}


	for _, c := range s.selects {
		if c.tokenCode == in.TableToken {
			c.chanMap[in.AuthRequest.Token] = make(chan SelectionContainer)
		}
	}

	is_pay_chan := false;
	for _, chans := range s.pays {
		if chans.tokenCode == in.TableToken {
			is_pay_chan = true
			break
		}
	}
	if !is_pay_chan {
		s.pays = append(s.pays, ItemPayNotificationChan{tokenCode: in.TableToken, tokenPays: make(chan ItemPayNotification)})
	}


	return OIR, nil
}

func (s* Server) OrderPay(ctx context.Context, in *OrderPayRequest) (*OrderPayResponse, error) {
	OPR, err := DBPayItem(in)
	if err != nil {
		return nil, err
	}

	fname, lname, err := DBAuthTokenToFirstLastName(in.AuthRequest.Token);
	if err != nil {
		return nil, err
	}
	cont := ItemPayNotification{Id: in.ItemPay.Id, Split: in.ItemPay.Split, PaidByFname: fname, PaidByLname: lname}
	for _, c := range s.pays {
		if c.tokenCode == in.ItemPay.TokenCode {
			c.tokenPays <- cont
		}

	}

	return OPR, nil
}
func (s *Server) ItemPaySubscribe(in *ItemPaySubscribeRequest, stream CPUser_ItemPaySubscribeServer) error {


	for _, c := range s.pays {
		if c.tokenCode == in.TokenCode {
			////send cache of clicks
			//for _, cachedCont := range c.selectCache {
			//	log.Print("Trying to send cache of selects")
			//	if err := stream.Send(&cachedCont); err != nil {
			//		log.Printf("Stream connection failed: %v", err)
			//		return nil
			//	}
			//}


			for {
				log.Print("Trying to sub")
				cont := <-c.tokenPays
				log.Printf("Got this from chan: %v", cont)

				//@HACK: dont spam client cuz client isnt buffered :(

				if err := stream.Send(&cont); err != nil {
					c.tokenPays <- cont
					log.Printf("Stream connection failed: %v", err)
					return nil
				}
			}

		}
	}
	return nil
}



// Selections

func RemoveFromSelectCacheIfFalse(sc []SelectionContainer, c SelectionContainer) []SelectionContainer {
	j := 0
	var tmpSc []SelectionContainer
	for _, n := range sc {
		//@TODO: change to phonenumber
		if n.Fname != c.Fname && n.Lname != c.Lname && n.ItemId != c.ItemId {
			copied := n
			sc[j] = n
			tmpSc = append(tmpSc, copied)
			j++;
		}

	}
	return tmpSc
}

func (s *Server) SelectionClick(ctx context.Context, in *SelectionRequest) (*emptypb.Empty, error) {
	//@TODO: DB selection
	//err := DBSelectionClick(in);
	//if err != nil {
	//	return err
	//}

	// make uuid so we can have some error checking stuff kinda to resend
	uuid, err := random(12)
	if err != nil {
		return &empty.Empty{}, status.Errorf(codes.Internal, "UUID Creation Error: %v")
	}

	fname, lname, err := DBAuthTokenToFirstLastName(in.AuthRequest.Token);
	if err != nil {
		return &empty.Empty{}, err
	}
	phone, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return &empty.Empty{}, err
	}

	cont := SelectionContainer{Fname: fname, Lname: lname, ItemId: in.Id, IsSplit: in.IsSplit, IsSelected: in.IsSelected, Phone: phone}
	cacheCont := cont

	log.Printf("Got from client: %v\n", cont)
	for i, c := range s.selects {
		if c.tokenCode == in.TokenCode {

			if cont.IsSelected {
				s.selects[i].selectCache = append(s.selects[i].selectCache, cacheCont)
			} else if !cont.IsSelected {
				s.selects[i].selectCache = RemoveFromSelectCacheIfFalse(s.selects[i].selectCache, cacheCont)
			}
			for _, element := range c.chanMap {
				element <- cont
			}
			break
		}
	}

	return &empty.Empty{}, nil

}

func (s *Server) SelectionInitial(ctx context.Context, in *SelectionCurrentUsersRequest) (*SelContArray, error) {
	for _, c := range s.selects {
		if c.tokenCode == in.TokenCode {
			contSlice := make([]*SelectionContainer, len(c.selectCache))
			for i := range contSlice {
				contSlice[i] = &c.selectCache[i]
			}
			return &SelContArray{Cont: contSlice}, nil
		}
	}
	return &SelContArray{}, status.Errorf(codes.NotFound, "Could not find token code");
}



func (s *Server) SelectionSubscribe(in *SelectionCurrentUsersRequest, stream CPUser_SelectionSubscribeServer) error {
	log.Print("Trying to sub")
	//@TODO: srv context to stop listening


	for i, c := range s.selects {
		if c.tokenCode == in.TokenCode {
			////send cache of clicks
			//for _, cachedCont := range c.selectCache {
			//	log.Print("Trying to send cache of selects")
			//	if err := stream.Send(&cachedCont); err != nil {
			//		log.Printf("Stream connection failed: %v", err)
			//		return nil
			//	}
			//}

			for {
				//log.Printf("notifchan: %\n", s.selects[i])
				cont := <-s.selects[i].chanMap[in.AuthRequest.Token]
				log.Printf("Got this from chan: %v\n", cont)

				//@HACK: dont spam client cuz client isnt buffered :(


				if err := stream.Send(&cont); err != nil {
					s.selects[i].chanMap[in.AuthRequest.Token] <- cont

					log.Printf("Stream connection failed: %v", err)
					return nil
				}
				//// listen for bounceback
				//for {
				//	req, err := stream.Recv()
				//	if err != nil {
				//		log.Println(err)
				//		return status.Errorf( codes.DataLoss,"Error receiving bounceback request: %v", err)
				//	}
				//	if req.LastUuid != cont.Uuid {
				//		if err := stream.Send(&cont); err != nil {
				//			c.tokenSelects <- cont
				//
				//			log.Printf("Stream connection failed: %v", err)
				//			return nil
				//		}
				//	}
				//	break
				//
				//}
			}

		}

	}



	return nil;
}




// Simple Ping -> Pong  sanity check
func (s *Server) Ping(ctx context.Context, in *emptypb.Empty) (*PongResponse, error) {
	log.Printf("Ping Ponged");
	return &PongResponse{PongMessage: "Pong!"}, nil
}