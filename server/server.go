package server

import (
	"github.com/fasthttp/router"
)

type BetaGoServer struct {
}

func (b *BetaGoServer) Start() {
	r := router.New()
}
