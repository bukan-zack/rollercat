package command

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func init() {
	Handlers["rollercat"] = rollercat
}

func rollercat(ch ssh.Channel, term *term.Terminal, args []string) {
	term.Write([]byte("https://github.com/zekflare/rollercat\n"))
}