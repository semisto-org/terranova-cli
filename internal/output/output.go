// Package output — la sortie est stylée au terminal et JSON dès qu'elle est
// redirigée (ISC-382). --json rend l'enveloppe {ok, data, summary, breadcrumbs},
// --quiet les données nues, --md une sortie portable, --ids-only / --count /
// --jq les raffinements d'agent (ISC-383).
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/semisto-org/terranova-cli/internal/cli"
	"golang.org/x/term"
)

// IsTTY dit si stdout est un terminal.
func IsTTY() bool { return term.IsTerminal(int(os.Stdout.Fd())) }

// Print rend un Result selon les drapeaux. C'est l'UNIQUE sortie de données du
// binaire — aucune commande n'imprime elle-même.
func Print(f cli.GlobalFlags, res *cli.Result) error {
	if res == nil {
		return nil
	}
	data := res.Data

	if f.JQ != "" {
		filtered, err := applyJQ(f.JQ, data)
		if err != nil {
			return err
		}
		data = filtered
	}

	switch {
	case f.Count:
		fmt.Println(countOf(data))
		return nil
	case f.IDsOnly:
		for _, id := range idsOf(data) {
			fmt.Println(id)
		}
		return nil
	case f.Quiet || f.Agent && !f.JSON:
		return printJSON(data)
	case f.JSON || !IsTTY():
		envelope := map[string]any{"ok": true, "data": data}
		if res.Summary != "" {
			envelope["summary"] = res.Summary
		}
		if len(res.Crumbs) > 0 {
			envelope["breadcrumbs"] = res.Crumbs
		}
		return printJSON(envelope)
	case f.MD:
		printMD(res, data)
		return nil
	default:
		printStyled(res, data)
		return nil
	}
}

// PrintError rend une erreur en enveloppe {ok:false, error:…} quand la sortie
// est JSON, sinon en une ligne sur stderr.
func PrintError(f cli.GlobalFlags, msg string, detail any) {
	if f.JSON || f.Agent || f.Quiet || !IsTTY() {
		envelope := map[string]any{"ok": false, "error": msg}
		if detail != nil {
			envelope["detail"] = detail
		}
		raw, _ := json.Marshal(envelope)
		fmt.Fprintln(os.Stderr, string(raw))
		return
	}
	fmt.Fprintln(os.Stderr, "✗ "+msg)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func printStyled(res *cli.Result, data any) {
	if len(res.Rows) > 0 {
		printTable(res.Headers, res.Rows)
	} else {
		_ = printJSON(data)
	}
	if res.Summary != "" {
		fmt.Println("— " + res.Summary)
	}
	for _, c := range res.Crumbs {
		fmt.Printf("  ↪ %s : %s\n", c.Action, c.Cmd)
	}
}

func printMD(res *cli.Result, data any) {
	if len(res.Rows) > 0 {
		fmt.Println("| " + strings.Join(res.Headers, " | ") + " |")
		seps := make([]string, len(res.Headers))
		for i := range seps {
			seps[i] = "---"
		}
		fmt.Println("| " + strings.Join(seps, " | ") + " |")
		for _, row := range res.Rows {
			fmt.Println("| " + strings.Join(row, " | ") + " |")
		}
	} else {
		raw, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println("```json")
		fmt.Println(string(raw))
		fmt.Println("```")
	}
	if res.Summary != "" {
		fmt.Println("\n> " + res.Summary)
	}
}

func printTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	line := func(cells []string) {
		parts := make([]string, len(cells))
		for i, cell := range cells {
			if i < len(widths) {
				parts[i] = fmt.Sprintf("%-*s", widths[i], cell)
			}
		}
		fmt.Println(strings.TrimRight(strings.Join(parts, "  "), " "))
	}
	line(headers)
	line(rowsOfDashes(widths))
	for _, row := range rows {
		line(row)
	}
}

func rowsOfDashes(widths []int) []string {
	out := make([]string, len(widths))
	for i, w := range widths {
		out[i] = strings.Repeat("─", w)
	}
	return out
}

func applyJQ(filter string, data any) (any, error) {
	q, err := gojq.Parse(filter)
	if err != nil {
		return nil, fmt.Errorf("--jq : filtre invalide : %w", err)
	}
	// gojq exige des types JSON purs : on repasse par un round-trip.
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var pure any
	if err := json.Unmarshal(raw, &pure); err != nil {
		return nil, err
	}
	var results []any
	iter := q.Run(pure)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return nil, err
		}
		results = append(results, v)
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// idsOf extrait les id d'une liste (ou d'un objet aux clés plurielles).
func idsOf(data any) []string {
	var ids []string
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			if id, ok := t["id"]; ok {
				ids = append(ids, fmt.Sprintf("%v", id))
				return
			}
			for _, sub := range t {
				walk(sub)
			}
		case []any:
			for _, item := range t {
				walk(item)
			}
		}
	}
	raw, _ := json.Marshal(data)
	var pure any
	_ = json.Unmarshal(raw, &pure)
	walk(pure)
	return ids
}

func countOf(data any) int {
	raw, _ := json.Marshal(data)
	var pure any
	_ = json.Unmarshal(raw, &pure)
	switch t := pure.(type) {
	case []any:
		return len(t)
	case map[string]any:
		for _, v := range t {
			if list, ok := v.([]any); ok {
				return len(list)
			}
		}
	}
	return 1
}
