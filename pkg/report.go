package pkg

import "go/token"

type Report struct {
	Pos     token.Pos
	Length  int
	Message string
}
