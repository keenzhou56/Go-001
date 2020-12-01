package service

import (
	"Go-000/dao"

	"database/sql"

	"github.com/pkg/errors"
)

type Service struct {
	dao *dao.Dao
}

func NewService() *Service {
	return &Service{dao.NewDao()}
}

func (s *Service) GetUserById(id int) (u dao.User, err error) {
	s = NewService()
	u, err = s.dao.FindUserById(id)

	if errors.Is(err, sql.ErrNoRows) {
		return u, nil
	}

	if err != nil {
		return u, errors.Wrapf(err, "service GetUserById(%d)", id)
	}

	return u, nil
}
