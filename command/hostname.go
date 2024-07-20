package command

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func init() {
	Handlers["hostname"] = hostname
}

func hostname(ch ssh.Channel, term *term.Terminal, args []string) {
	term.Write([]byte("rollercat\n"))
}