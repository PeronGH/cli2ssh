package main

import (
	"flag"
	"os"
	"os/exec"

	"github.com/PeronGH/cli2ssh/internal/args"
	"github.com/PeronGH/cli2ssh/internal/path"
	"github.com/PeronGH/cli2ssh/internal/server"
	"github.com/PeronGH/cli2ssh/internal/utils"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func main() {
	command := flag.String("c", "", "Set the command to run for each SSH session.")
	var env args.ArrayArg
	flag.Var(&env, "e", "Set environment variables for each SSH session.")
	useOsEnv := flag.Bool("os-env", false, "Use the OS environment variables for the command, ignoring env passed by user.")
	host := flag.String("h", "localhost", "Set the host for the server.")
	port := flag.String("p", "2222", "Set the port for the server.")
	hostKeyPath := flag.String("k", path.GetDefaultHostKeyPath(), "Set the path to the host key.")

	flag.Parse()

	if *command == "" {
		log.Fatal("No command provided.")
	}

	srv, err := server.CreateServer(server.CreateServerOptions{
		CommandProvider: func(s ssh.Session) *exec.Cmd {
			argSession := args.NewSession(s)
			fmtCmd := argSession.FormatArg(*command)
			fmtEnv := argSession.FormatArgs(env)

			cmd := exec.CommandContext(s.Context(), "sh", "-c", fmtCmd)
			if *useOsEnv {
				cmd.Env = os.Environ()
			} else {
				cmd.Env = s.Environ()
			}
			cmd.Env = append(cmd.Env, fmtEnv...)
			return cmd
		},

		Host:        *host,
		Port:        *port,
		HostKeyPath: *hostKeyPath,

		PublicKeyAuth: func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Info("Public key auth", "remote", ctx.RemoteAddr(), "user", ctx.User(), "key", utils.StringifyPublicKey(key))
			return true
		},
		PasswordAuth: func(ctx ssh.Context, password string) bool {
			log.Info("Password auth", "remote", ctx.RemoteAddr(), "user", ctx.User(), "password", password)
			return true
		},
	})

	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	log.Info("Starting server...", "address", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
