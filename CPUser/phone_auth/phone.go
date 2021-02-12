/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

package phone_auth

import (
	"crypto/rand"
	"fmt"
)
import "github.com/kevinburke/twilio-go"

const (
	sid = "ACc897f7a19949cec0bef9acc4301b5048"
	token = "a4d9f7fd8b9689df9a3b1ca499fa375a"
)

var client = twilio.NewClient(sid, token, nil)


const chars = "0123456789"
const TOKEN_LENGTH = 6

func random(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes), nil
}
func SendTextVerification(phone_num string) (string, error) {
	fmt.Println("GOT VERIFICATON REQUEST FOR: " + phone_num)

	GeneratedSMSToken, err := random(TOKEN_LENGTH)
	if (err != nil) {
		return "", err
	}

	_, err = client.Messages.SendMessage("+12314409896",
		"+1"+phone_num,
		"CheckPlease Authentication Code: " +GeneratedSMSToken,
		nil)
	if (err != nil) {
		return "", err
	}


	return GeneratedSMSToken, nil
}