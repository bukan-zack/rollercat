package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"

	"log/slog"

	"github.com/fatih/color"
	"github.com/zekflare/rollercat/command"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {
	logger := slog.Default()

	if _, err := os.Stat(".rollercat/id_ed25519"); os.IsNotExist(err) {
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			logger.Error("could not generate ed25519 private key", slog.Any("err", err))
			os.Exit(1)
		}

		if err := os.MkdirAll(".rollercat", 0o755); err != nil {
			logger.Error("could not create rollercat directory", slog.Any("err", err))
			os.Exit(1)
		}

		f, err := os.OpenFile(".rollercat/id_ed25519", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			logger.Error("could not open rollercat private key", slog.Any("err", err))
			os.Exit(1)
		}
		defer f.Close()
	
		b, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			logger.Error("could not marshal rollercat private key", slog.Any("err", err))
			os.Exit(1)
		}

		if err := pem.Encode(f, &pem.Block{Type: "PRIVATE KEY", Bytes: b}); err != nil {
			logger.Error("could not write rollercat private key", slog.Any("err", err))
			os.Exit(1)
		}
	}

	pkb, err := os.ReadFile(".rollercat/id_ed25519")
	if err != nil {
		logger.Error("could not read rollercat private key", slog.Any("err", err))
		os.Exit(1)
	}

	pk, err := ssh.ParsePrivateKey(pkb)
	if err != nil {
		logger.Error("could not parse rollercat private key", slog.Any("err", err))
		os.Exit(1)
	}

	sshCfg := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	sshCfg.AddHostKey(pk)

	listener, err := net.Listen("tcp", "0.0.0.0:6969")
	if err != nil {
		logger.Error("could not listen tcp", slog.Any("err", err))
		os.Exit(1)
	}

	for {
		if nConn, _ := listener.Accept(); nConn != nil {
			go func() {
				if err := acceptInbound(nConn, sshCfg); err != nil {
					logger.Error("could not accept incoming connection", slog.Any("err", err))
				}
			}()
		}
	}
}

func acceptInbound(nConn net.Conn, cfg *ssh.ServerConfig) error {
	conn, chans, reqs, err := ssh.NewServerConn(nConn, cfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	go ssh.DiscardRequests(reqs)

	for newCh := range chans {
		if newCh.ChannelType() != "session" {
			newCh.Reject(ssh.UnknownChannelType, "unknown channel type")
			
			continue
		}

		ch, reqs, err := newCh.Accept()
		if err != nil {
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				req.Reply(req.Type == "shell" || req.Type == "pty-req", nil)
			}
		}(reqs)

		term := term.NewTerminal(ch, prompt(conn.User()))
		term.Write([]byte(`
    ____        ____                     __ 
   / __ \____  / / /__  ______________ _/ /_
  / /_/ / __ \/ / / _ \/ ___/ ___/ __ '/ __/
 / _, _/ /_/ / / /  __/ /  / /__/ /_/ / /_  
/_/ |_|\____/_/_/\___/_/   \___/\__,_/\__/  
                                            
Repository: https://github.com/zekflare/rollercat

`))

		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}

			args := strings.Split(line, " ")
			if args[0] == "" {
				continue
			}

			if handler, ok := command.Handlers[args[0]]; ok {
				handler(ch, term, args)

				continue
			}

			term.Write([]byte(fmt.Sprintf("zsh: command not found: %s\n", args[0])))
		}
	}

	return nil
}

func prompt(user string) string {
	var b strings.Builder

	color.New(color.FgGreen, color.Bold).Fprint(&b, user+"@rollercat")
	fmt.Fprintf(&b, ":")
	color.New(color.FgBlue, color.Bold).Fprint(&b, "~")
	fmt.Fprintf(&b, "$ ")

	return b.String()
}