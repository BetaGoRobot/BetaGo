package server

import (
	"github.com/fasthttp/router"
)

type BetaGoServer struct{}

func (b *BetaGoServer) Start() {
	_ = router.New()
}
