package cli

import (
	"fmt"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/msg"
)

// ParseGlobals sépare les drapeaux globaux des arguments de commande. Ils
// peuvent apparaître n'importe où sur la ligne (patron basecamp-cli).
func ParseGlobals(args []string) (GlobalFlags, []string, error) {
	f := GlobalFlags{}
	rest := []string{}
	i := 0
	for i < len(args) {
		a := args[i]
		next := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf(msg.UsageFlagNeedsValue, a)
			}
			i++
			return args[i], nil
		}
		switch a {
		case "--json", "-j":
			f.JSON = true
		case "--quiet", "-q":
			f.Quiet = true
		case "--md", "-m":
			f.MD = true
		case "--agent":
			f.Agent = true
			f.JSON = true
			f.Quiet = true
		case "--ids-only":
			f.IDsOnly = true
		case "--count":
			f.Count = true
		case "--yes":
			f.Yes = true
		case "--help", "-h":
			f.Help = true
		case "-v":
			f.Verbose = 1
		case "-vv":
			f.Verbose = 2
		case "--jq":
			v, err := next()
			if err != nil {
				return f, nil, err
			}
			f.JQ = v
		case "--hub":
			v, err := next()
			if err != nil {
				return f, nil, err
			}
			f.Hub = v
		case "--project", "-p", "--in":
			v, err := next()
			if err != nil {
				return f, nil, err
			}
			f.Project = v
		case "--profile", "-P":
			v, err := next()
			if err != nil {
				return f, nil, err
			}
			f.Profile = v
		default:
			rest = append(rest, a)
		}
		i++
	}
	return f, rest, nil
}

// UsageError signale une erreur d'usage (code de sortie 2).
type UsageError struct{ Msg string }

func (e *UsageError) Error() string { return e.Msg }

func Usagef(format string, args ...any) error {
	return &UsageError{Msg: fmt.Sprintf(format, args...)}
}

// HelpText rend l'aide humaine (français) d'une commande ou de la racine.
func HelpText(path string, c *Command) string {
	var b strings.Builder
	if c == nil {
		b.WriteString(msg.HelpRootBanner)
		b.WriteString(msg.HelpRootUsage)
		groups := map[string][]string{}
		order := []string{}
		for _, cmd := range Registry {
			g := cmd.Group
			if g == "" {
				g = msg.GroupOther
			}
			if _, seen := groups[g]; !seen {
				order = append(order, g)
			}
			groups[g] = append(groups[g], fmt.Sprintf("  %-16s %s", cmd.Name, cmd.Summary))
		}
		for _, g := range order {
			b.WriteString(fmt.Sprintf(msg.HelpGroupHeader, g) + strings.Join(groups[g], "\n") + "\n\n")
		}
		b.WriteString(msg.HelpRootGlobalFlags1)
		b.WriteString(msg.HelpRootGlobalFlags2)
		b.WriteString(msg.HelpRootFooter)
		return b.String()
	}
	b.WriteString(fmt.Sprintf(msg.HelpCommandTitle, path, c.Summary))
	if c.ArgSpec != "" {
		b.WriteString(fmt.Sprintf(msg.HelpCommandUsage, path, c.ArgSpec))
	}
	if len(c.Flags) > 0 {
		b.WriteString(msg.HelpFlagsHeader)
		for _, fl := range c.Flags {
			name := "--" + fl.Name
			if fl.Short != "" {
				name += ", -" + fl.Short
			}
			if fl.Arg != "" {
				name += " <" + fl.Arg + ">"
			}
			b.WriteString(fmt.Sprintf("  %-28s %s\n", name, fl.Help))
		}
	}
	if len(c.Sub) > 0 {
		b.WriteString(msg.HelpSubcommandsHeader)
		for _, s := range c.Sub {
			b.WriteString(fmt.Sprintf("  %-16s %s\n", s.Name, s.Summary))
		}
	}
	if c.AgentHelp != "" {
		b.WriteString(msg.HelpNotesPrefix + c.AgentHelp + "\n")
	}
	return b.String()
}

// AgentHelp rend l'aide structurée --help --agent (ISC-385).
func AgentHelpFor(path string, c *Command) map[string]any {
	if c == nil {
		cmds := []map[string]any{}
		Walk(func(p string, cmd *Command) {
			cmds = append(cmds, map[string]any{"cmd": p, "summary": cmd.Summary})
		})
		return map[string]any{"name": "terranova", "commands": cmds}
	}
	return map[string]any{
		"name":        "terranova " + path,
		"summary":     c.Summary,
		"args":        c.ArgSpec,
		"flags":       c.Flags,
		"subcommands": subNames(c),
		"notes":       c.AgentHelp,
		"api_ops":     c.APIOps,
	}
}

func subNames(c *Command) []string {
	names := make([]string, 0, len(c.Sub))
	for _, s := range c.Sub {
		names = append(names, s.Name)
	}
	return names
}

// FlagValue lit un drapeau propre à la commande dans ses arguments restants
// (`--due 2026-09-01`), et rend les arguments épurés.
func FlagValue(args []string, name string) (string, []string) {
	out := []string{}
	val := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--"+name && i+1 < len(args) {
			val = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--"+name+"=") {
			val = strings.TrimPrefix(args[i], "--"+name+"=")
			continue
		}
		out = append(out, args[i])
	}
	return val, out
}

// FlagBool lit un drapeau booléen propre à la commande.
func FlagBool(args []string, name string) (bool, []string) {
	out := []string{}
	found := false
	for _, a := range args {
		if a == "--"+name {
			found = true
			continue
		}
		out = append(out, a)
	}
	return found, out
}
