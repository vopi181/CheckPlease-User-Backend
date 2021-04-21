package CPUser

import (
	"context"
	"testing"
)


func TestCreateUser(t *testing.T) {
	s := Server{}

	req := &CreateUserRequest{Phone: "8475551234", Fname: "TestFname", Lname: "TestLname"}

	resp, err := s.CreateUser(context.Background(), req)
	if err != nil {
		t.Errorf("TestCreateUser(%v,%v,%v) got unexpected error %v", req.Phone, req.Fname, req.Lname, err)
	}
	if resp.AuthToken != "" {
		t.Errorf("TestCreateUser did not returned expected  %v", err)
	}
}
