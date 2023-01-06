package auth

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/stirboy/jh/pkg/config"
)

func CheckAuth(cfg config.Config) bool {
	token, _ := cfg.AuthToken()
	if token != "" {
		return true
	}

	return strings.TrimSpace(token) != ""
}

func IsAuthEnabled(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return false
	}

	// going up in command chain to check
	// if any parent requires auth
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Annotations != nil && c.Annotations["skip_auth"] == "true" {
			return false
		}
	}

	return true
}

func DisableAuthCheck(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}

	cmd.Annotations["skip_auth"] = "true"
}
