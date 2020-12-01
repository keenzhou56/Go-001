package main

import (
	"Go-000/week02/service"
	"fmt"
)

func main() {
	s := service.NewService()
	u, err := s.GetUserById(1)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	fmt.Printf("%+v\n", u)
}
