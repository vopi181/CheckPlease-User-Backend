/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

package CPUser

import (
	"crypto/rand"
	"log"
)

//@TODO: BIG TODO ADD ACTUAL VETTED AUTH (maybe tho)
const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const TOKEN_LENGTH = 24

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

func GenerateAuthToken(phone string) string {
	tok, err := random(TOKEN_LENGTH)
	if(err != nil) {
		log.Fatal(err)
	}
	return tok
}