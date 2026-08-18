// terranova — le compagnon en ligne de commande de Terranova.
// Codes de sortie (ISC-379) : 0 succès · 1 erreur d'exécution · 2 usage ·
// 3 authentification · 4 permission · 5 introuvable.
package main

import (
	"fmt"
	"os"

	"github.com/semisto-org/terranova-cli/internal/api"
	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/commands"
	"github.com/semisto-org/terranova-cli/internal/config"
	"github.com/semisto-org/terranova-cli/internal/msg"
	"github.com/semisto-org/terranova-cli/internal/output"
)

// Version est posée par -ldflags au release (ISC-369).
var Version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	commands.Version = Version

	flags, rest, err := cli.ParseGlobals(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		output.PrintError(flags, err.Error(), nil)
		return 1
	}

	ctx := &cli.Ctx{
		Flags:   flags,
		Config:  cfg,
		Profile: cfg.ResolveProfile(flags.Profile),
		Version: Version,
		IsTTY:   output.IsTTY(),
	}

	cmd, cmdArgs := cli.Find(rest)

	// Aide : racine, commande, ou structurée pour agents.
	if cmd == nil && len(rest) > 0 && !flags.Help {
		output.PrintError(flags, fmt.Sprintf(msg.ErrUnknownCommand, rest[0]), nil)
		return 2
	}
	if flags.Help || cmd == nil || (cmd.Run == nil && len(cmdArgs) == 0) {
		path := pathOf(rest, cmdArgs)
		if flags.Agent {
			_ = output.Print(flags, &cli.Result{Data: cli.AgentHelpFor(path, cmd)})
			return 0
		}
		fmt.Print(cli.HelpText(path, cmd))
		if cmd == nil || cmd.Run == nil {
			return 0
		}
		return 0
	}
	if cmd.Run == nil {
		output.PrintError(flags, fmt.Sprintf(msg.ErrUnknownSubcommand, cmdArgs[0], pathOf(rest, cmdArgs)), nil)
		return 2
	}
	if len(cmdArgs) < cmd.MinArgs {
		output.PrintError(flags, fmt.Sprintf(msg.UsageCommand, pathOf(rest, cmdArgs), cmd.ArgSpec), nil)
		return 2
	}

	res, err := cmd.Run(ctx, cmdArgs)
	if err != nil {
		return renderError(flags, err)
	}
	if err := output.Print(flags, res); err != nil {
		return renderError(flags, err)
	}
	return 0
}

func pathOf(rest, cmdArgs []string) string {
	n := len(rest) - len(cmdArgs)
	if n <= 0 {
		return ""
	}
	path := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			path += " "
		}
		path += rest[i]
	}
	return path
}

func renderError(flags cli.GlobalFlags, err error) int {
	if usage, ok := err.(*cli.UsageError); ok {
		output.PrintError(flags, usage.Msg, nil)
		return 2
	}
	if apiErr, ok := err.(*api.Error); ok {
		var detail any
		if len(apiErr.Body) > 0 {
			detail = apiErr.Body
		}
		output.PrintError(flags, apiErr.Error(), detail)
		switch apiErr.Status {
		case 401:
			return 3
		case 403:
			return 4
		case 404:
			return 5
		default:
			return 1
		}
	}
	output.PrintError(flags, err.Error(), nil)
	return 1
}
