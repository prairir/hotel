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

	port := os.Getenv("HOTEL_PORT")
	if port == "" {
		zlog.Info("`HOTEL_PORT` is empty, defaulting to 2222")
		port = "2222"
	}

	port = ":" + port

	hostKeyPath := os.Getenv("HOTEL_HOST_KEY_PATH")
	if hostKeyPath == "" {
		zlog.Info("`HOTEL_HOST_KEY_PATH` is empty, generating Host Key")

		zlog.Info("starting ssh server", "port", port)
		err = ssh.ListenAndServe(port,
			handler.Handler(dock, zlog),
			ssh.PasswordAuth(password.Handler(zlog)))
	} else {
		zlog.Info("starting ssh server", "port", port)
		err = ssh.ListenAndServe(port,
			handler.Handler(dock, zlog),
			ssh.PasswordAuth(password.Handler(zlog)),
			ssh.HostKeyFile(hostKeyPath),
		)
	}

	zlog.Info("starting ssh server", "port", port)
	err = ssh.ListenAndServe(port,
		handler.Handler(dock, zlog),
		ssh.PasswordAuth(password.Handler(zlog)))
	zlog.Error(err, "ssh.ListenAndServer")
}
