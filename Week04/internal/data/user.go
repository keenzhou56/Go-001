package data

import (
	"Week04/internal/biz"
	"log"
)

var _ biz.UserRepo = new(userRepo)

const (
	userID = 100
)

func NewUserRepo() biz.UserRepo {
	return &userRepo{}
}

type userRepo struct{}

func (r *userRepo) Save(u *biz.User) int32 {
	log.Printf("save username: %s, password: %s, id: %d", u.Username, u.Password, userID)
	return userID
}
