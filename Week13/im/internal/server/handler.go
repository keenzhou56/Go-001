package server

import (
	"im/pkg/config"
)

func (s *Server) RegisterHandler() {
	s.InitRouter()
	s.AddRouter(config.ImLogin, s.LoginHandler)
	s.AddRouter(config.ImChatTestReply, s.ApitestHandler)
	s.AddRouter(config.ImKickUser, s.KickUserHandler)
	s.AddRouter(config.ImKickAll, s.KickAllHandler)
	s.AddRouter(config.ImStat, s.StatHandler)
}
