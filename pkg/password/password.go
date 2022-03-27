package password

import (
	"fmt"

	"github.com/gliderlabs/ssh"
	"github.com/go-logr/logr"
	"github.com/msteinert/pam"
)

func Handler(log logr.Logger) func(ctx ssh.Context, password string) bool {
	return func(ctx ssh.Context, password string) bool {
		pm, err := pam.StartFunc("", ctx.User(), func(s pam.Style, msg string) (string, error) {
			return password, nil
		})
		if err != nil {
			log.Error(fmt.Errorf("password.Handler: %w", err), "StartFunc")
			return false
		}

		err = pm.Authenticate(0)
		if err != nil {
			log.Error(fmt.Errorf("password.Handler: %w", err), "Authenticate")
			return false
		}

		err = pm.AcctMgmt(pam.Silent)
		if err != nil {
			log.Error(fmt.Errorf("password.Handler: %w", err), "Authenticate")
			return false
		}

		ctx.SetValue("password", password)

		return true
	}
}
