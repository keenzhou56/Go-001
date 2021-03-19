package server

import (
	"context"
	"im/api/logic"
)

// Receive receive a message.
func (s *Server) Receive(ctx context.Context, userID int64, p *logic.Proto) (err error) {
	_, err = s.rpcClient.Receive(ctx, &logic.ReceiveReq{UserID: userID, Proto: p})
	return
}
