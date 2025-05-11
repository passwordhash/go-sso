package models

type User struct {
	UUID     string
	Email    string
	PassHash []byte
}
