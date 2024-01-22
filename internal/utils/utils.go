package utils

import (
	"encoding/base64"
	"math/rand"
	"strconv"
)

const (
	maxRandNumber = 100000
)

func GenerateRandomString() string {
	return string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(rand.Intn(maxRandNumber)))))
}
