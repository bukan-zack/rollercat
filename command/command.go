package command

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Handler func(ch ssh.Channel, term *term.Terminal, args []string)

var Handlers map[string]Handler = make(map[string]Handler, 0)