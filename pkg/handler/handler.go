package handler

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
	"github.com/prairir/hotel/pkg/docker"

	"github.com/docker/docker/api/types/container"
	"github.com/gliderlabs/ssh"
)

func Handler(dock *docker.Dock, log logr.Logger) func(ssh.Session) {
	return func(sess ssh.Session) {
		log.Info("connection started", "addr", sess.RemoteAddr())

		// pull password out of context
		pass := sess.Context().Value("password")

		name := sess.User()

		err := dock.BuildContainer(name, sess.User(), pass, ".")
		if err != nil {
			log.Error(fmt.Errorf("handler.Handler: %w", err), "build container")
			io.WriteString(sess, fmt.Errorf("ERROR handler.Handler: %w\n\nEXITING", err).Error())
			sess.Exit(1)
		}

		_, _, isTty := sess.Pty()

		config := container.Config{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			Tty:          isTty,
			Image:        name,
		}

		id, err := dock.RunContainer(name, config)
		if err != nil {
			log.Error(fmt.Errorf("handler.Handler: %w", err), "run container")
			io.WriteString(sess, fmt.Errorf("ERROR handler.Handler: %w\n\nEXITING", err).Error())
			sess.Exit(1)
		}

		waiter, err := dock.AttachContainer(id, sess, sess)
		if err != nil {
			log.Error(fmt.Errorf("handler.Handler: %w", err), "attach container")
			io.WriteString(sess, fmt.Errorf("ERROR handler.Handler: %w\n\nEXITING", err).Error())
			sess.Exit(1)
		}
		defer waiter.Close()

		err = dock.WaitContainer(id)
		if err != nil {
			log.Error(fmt.Errorf("handler.Handler: %w", err), "wait container")
			io.WriteString(sess, fmt.Errorf("ERROR handler.Handler: %w\n\nEXITING", err).Error())
			sess.Exit(1)
		}

		log.Info("connection leaving", "addr", sess.RemoteAddr())
	}

}
