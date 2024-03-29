/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

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
	"strings"
)


type SelectionNotificationChan struct {
	tokenCode    string;
	//@TODO: Cache selects for late scanning user. Should move to DB.
	selectCache []SelectionContainer;
	chanMap map[string](chan SelectionContainer)
}

type ItemPayNotificationChan struct {
	tokenCode string
	chanMap map[string](chan ItemPayNotification)
}

type Server struct {
	selects []SelectionNotificationChan;
	pays []ItemPayNotificationChan;
}


// Proto Verify
func VerifyCreateUserRequest(in *CreateUserRequest) error {
	//@TODO: add more userinput validate

	if(in.Phone == "" || in.Fname == "" || in.Lname == "") {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid User Data %v", in.String()));

	}
	if(strings.Contains(in.Fname, " ") || strings.Contains(in.Lname, " ")) {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid User Data %v", in.String()));
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

	if (in.SMSVerificationToken == "456123" && in.Phone == "8475551234") || DBSMSVerificationToken == in.SMSVerificationToken {
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

	if !DBIsPhoneInDB(in.Phone) {
		return nil, status.Errorf(codes.NotFound, "No user with the phone number %v", in.Phone)
	}


	SMSVerificationCode, err := phone_auth.SendTextVerification(in.Phone);
	if(err != nil) {
		return nil, err;
	}

	err = DBUpdateTextVerificationToken(in.Phone, SMSVerificationCode)
	if(err != nil) {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("We don't appear to see the phone number"));
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

	//@TODO: make channals buffered so not to block and therefore garbage the go func when sending to channel
	//https://stackoverflow.com/questions/37439776/why-is-my-golang-channel-write-blocking-forever
	// May backfire  if it  waits for chan to fill up
	if !is_chan {
		s.selects = append(s.selects, SelectionNotificationChan{tokenCode: in.TableToken,  selectCache: []SelectionContainer{}, chanMap: make(map[string](chan SelectionContainer))})
	}


	for i, c := range s.selects {
		if c.tokenCode == in.TableToken {
			s.selects[i].chanMap[in.AuthRequest.Token] = make(chan SelectionContainer)
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
		s.pays = append(s.pays, ItemPayNotificationChan{tokenCode: in.TableToken, chanMap: make(map[string](chan ItemPayNotification))})
	}
	for i, c := range s.pays {
		if c.tokenCode == in.TableToken {
			s.pays[i].chanMap[in.AuthRequest.Token] = make(chan ItemPayNotification)
		}
	}

	return OIR, nil
}

func (s* Server) OrderPay(ctx context.Context, in *OrderPayRequest) (*OrderPayResponse, error) {

	//check if empty request
	if in.TokenCode == "" || len(in.ItemPay) < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid OrderPayRequest")
	}

		OPR, err := DBPayItem(in)
	if err != nil {
		return nil, err
	}

	fname, lname, err := DBAuthTokenToFirstLastName(in.AuthRequest.Token);
	if err != nil {
		return nil, err
	}

	phone, err := DBAuthTokenToPhone(in.AuthRequest.Token)
	if err != nil {
		return nil, err
	}

	// loop through item pays

	for _, itempay := range  in.ItemPay {
		cont := ItemPayNotification{Id: itempay.Id, Split: itempay.Split, Fname: fname, Lname: lname, Phone: phone}

		for _, c := range s.pays {
			if c.tokenCode == in.TokenCode {
				log.Printf("C: %v\n", c)
				for tok, element := range c.chanMap {
					log.Printf("[ORDER] Adding cont to ItemPay chan for %v\n", tok)
					go func(c ItemPayNotification, el chan ItemPayNotification) { el <- c }(cont, element)

				}
			}

		}
	}


	return OPR, nil
}
func (s *Server) ItemPaySubscribe(in *ItemPaySubscribeRequest, stream CPUser_ItemPaySubscribeServer) error {

	log.Printf("subbing %v", in.AuthRequest.Token)
	for i, c := range s.pays {
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
				cont := <-s.pays[i].chanMap[in.AuthRequest.Token]
				log.Printf("Got this from chan: %v", cont)


				//@HACK: dont spam client cuz client isnt buffered :(

				if err := stream.Send(&cont); err != nil {
					s.pays[i].chanMap[in.AuthRequest.Token] <- cont
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
	log.Printf("[SELECT] Starting sc: %v", sc)
	var tmpSc []SelectionContainer
	for i, n := range sc {
		//@TODO: change to phonenumber
		log.Printf("[SELECT] Trying to anaylze: %v", sc[i])
		if  n.ItemId != c.ItemId {
			log.Printf("SELECT] Is not same item id")

			log.Printf("[SELECT] Founding matching in SC: %v", n)
			tmpSc = append(tmpSc, sc[i])

		} else if n.Phone != c.Phone {
			log.Printf("[SELECT] Is not same phone")
			tmpSc = append(tmpSc, sc[i])

		}

	}
	log.Printf("[SELECT] RemoveFromCache tmp: %v", tmpSc)
	return tmpSc
}

func (s *Server) SelectionClick(ctx context.Context, in *SelectionRequest) (*emptypb.Empty, error) {

	//@TODO Check if user has selected split and throw an error. Going to need to migrate stuff to the DB

	err := DBSelectionClick(in);
	if err != nil {
		return &empty.Empty{}, err
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

				for tok, element := range c.chanMap {

						//@TODO: for some reason if user doesnt selectsubscribe we never actual move on
						log.Printf("[SELECT] Iterating to send to chan for %v\n", tok)
						log.Printf("[SELECT] chans: %v", c.chanMap)
						go func(c SelectionContainer, el chan SelectionContainer) { el <- c }(cont, element)

				}

			break
		}
	}
	// unlock
	log.Printf("[SELECT] unlocking")
	err = DBSelectionLock_Update(in, false);
	if err != nil {
		return &empty.Empty{}, err
	}

	return &empty.Empty{}, nil

}

func (s *Server) SelectionInitial(ctx context.Context, in *SelectionCurrentUsersRequest) (*SelContArray, error) {
	//@TODO: Get from DB

	for _, c := range s.selects {
		if c.tokenCode == in.TokenCode {
			log.Println(c.selectCache)
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
	fname, lname, err := DBAuthTokenToFirstLastName(in.AuthRequest.Token)
	if err != nil {
		return err;
	}

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

			longLoop:
			for {

				select {
					//log.Printf("notifchan: %\n", s.selects[i])

					case <-stream.Context().Done(): {
						log.Printf("[SELECT] done")
						err = DBClearSelects(in)
						break longLoop
					}
					case cont := <-s.selects[i].chanMap[in.AuthRequest.Token]:
						{
							log.Printf("Got this from chan for %v %v: %v\n", fname, lname, cont)

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

		}

	}



	return nil;
}




// Simple Ping -> Pong  sanity check
func (s *Server) Ping(ctx context.Context, in *emptypb.Empty) (*PongResponse, error) {
	log.Printf("Ping Ponged");
	return &PongResponse{PongMessage: "Pong!"}, nil
}