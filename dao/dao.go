package dao

import (
	"database/sql"
	// "github.com/pkg/errors"
)

var (
	DB *sql.DB
)

type User struct {
	Id   int
	Name string
}

type Dao struct {
}

func NewDao() *Dao {
	return &Dao{}
}

func (d *Dao) FindUserById(userID int) (u User, err error) {
	err = sql.ErrNoRows
	return u, err
}
