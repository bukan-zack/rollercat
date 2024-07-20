package command

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func init() {
	Handlers["exit"] = exit
}

func exit(ch ssh.Channel, term *term.Terminal, args []string) {
	ch.Close()	
}