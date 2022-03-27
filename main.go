package main

import (
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/prairir/hotel/pkg/docker"
	"github.com/prairir/hotel/pkg/handler"
	"github.com/prairir/hotel/pkg/password"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

func main() {
	zl := zerolog.New(os.Stdout)

	zlr := zl.With().Logger()

	zlog := zerologr.New(&zlr)
	dock, err := docker.New(zlog)
	if err != nil {
		panic(err)
	}

	ssh.Handle(handler.Handler(dock, zlog))

	port := ":2222"

	zlog.Info("starting ssh server", "port", port)
	err = ssh.ListenAndServe(port,
		handler.Handler(dock, zlog),
		ssh.PasswordAuth(password.Handler(zlog)))
	zlog.Error(err, "ssh.ListenAndServer")
}
