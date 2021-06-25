package utils

import (
	"math/rand"
	"time"
)

func GetRandomString(index int, length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(index)])
	}
	return string(result)
}

func Map(s []string, f func(string) string) []string {
	r := make([]string, 0, len(s))
	for _, ss := range s {
		r = append(r, f(ss))
	}
	return r
}
