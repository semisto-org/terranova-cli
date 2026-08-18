package commands

import (
	"fmt"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// zz_* : greffé APRÈS l'enregistrement de `nurserio` par lenses.go (ordre
// d'initialisation alphabétique des fichiers Go — même raison que le chat).
func init() {
	orders := &cli.Command{
		Name: "orders", Summary: msg.HelpOrders,
		Sub: []*cli.Command{
			{
				Name: "list", Summary: msg.HelpOrdersList,
				Flags:  []cli.Flag{{Name: "status", Arg: "statut", Help: msg.FlagFiltreParEtat}},
				APIOps: []string{"GET /nurserio/orders"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					status, _ := cli.FlagValue(args, "status")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/nurserio/orders"
					if status != "" {
						path += "?status=" + status
					}
					var out struct {
						Orders []map[string]any `json:"orders"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, o := range out.Orders {
						rows = append(rows, []string{str(o["id"]), str(o["order_number"]), str(o["status"]), str(o["customer_name"]), str(o["total_cents"])})
					}
					return &cli.Result{Data: out.Orders, Headers: msg.HeadersOrders, Rows: rows,
						Summary: fmt.Sprintf(msg.ResOrderCount, len(out.Orders)),
						Crumbs:  []cli.Crumb{{Action: msg.CrumbPreparer, Cmd: "terranova nurserio orders process <id>"}}}, nil
				},
			},
			{
				Name: "show", Summary: msg.HelpOrdersShow, ArgSpec: "<id>", MinArgs: 1,
				APIOps: []string{"GET /nurserio/orders/{id}"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Get("/nurserio/orders/"+args[0], &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out}, nil
				},
			},
			{
				Name: "add", Summary: msg.HelpOrdersAdd,
				Flags: []cli.Flag{
					{Name: "line", Arg: "batch:qty", Help: msg.FlagOrdersAddLine},
					{Name: "customer", Arg: "nom", Help: msg.FlagOrdersAddCustomer},
					{Name: "email", Arg: "email", Help: msg.FlagOrdersAddEmail},
					{Name: "phone", Arg: "tél", Help: msg.FlagOrdersAddPhone},
					{Name: "member-id", Arg: "id", Help: msg.FlagOrdersAddMemberId},
					{Name: "pickup-site-id", Arg: "id", Help: msg.FlagOrdersAddPickupSiteId},
					{Name: "pickup-on", Arg: "date", Help: msg.FlagOrdersAddPickupOn},
					{Name: "notes", Arg: "texte", Help: msg.FlagNotes},
				},
				APIOps: []string{"POST /nurserio/orders"},
				Run:    runOrderAdd,
			},
			orderGesture("process", msg.HelpOrdersProcess),
			orderGesture("ready", msg.HelpOrdersReady),
			orderGesture("pickup", msg.HelpOrdersPickup),
			orderGesture("cancel", msg.HelpOrdersCancel),
		},
	}
	for _, cmd := range cli.Registry {
		if cmd.Name == "nurserio" {
			cmd.Sub = append(cmd.Sub, orders)
		}
	}
}

func orderGesture(name, summary string) *cli.Command {
	return &cli.Command{
		Name: name, Summary: summary, ArgSpec: "<id>", MinArgs: 1,
		APIOps: []string{"POST /nurserio/orders/{id}/" + name},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			var out map[string]any
			if err := client.Post("/nurserio/orders/"+args[0]+"/"+name, nil, &out); err != nil {
				return nil, err
			}
			summaryLine := summary
			if unservable, ok := out["unservable_lines"].([]any); ok && len(unservable) > 0 {
				summaryLine = fmt.Sprintf(msg.ResUnservableLines, len(unservable))
			}
			return &cli.Result{Data: out, Summary: summaryLine}, nil
		},
	}
}

func runOrderAdd(c *cli.Ctx, args []string) (*cli.Result, error) {
	lines := []map[string]any{}
	rest := args
	for {
		var line string
		line, rest = cli.FlagValue(rest, "line")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, cli.Usagef(msg.UsageOrderLineFormat, line)
		}
		lines = append(lines, map[string]any{"stock_batch_id": parts[0], "quantity": parts[1]})
	}
	if len(lines) == 0 {
		return nil, cli.Usagef(msg.UsageOrderLineRequired)
	}
	body := map[string]any{"lines": lines}
	for flag, param := range map[string]string{"customer": "customer_name", "email": "customer_email",
		"phone": "customer_phone", "member-id": "member_id", "pickup-site-id": "pickup_site_id",
		"pickup-on": "desired_pickup_on", "notes": "notes"} {
		v, r := cli.FlagValue(rest, flag)
		rest = r
		if v != "" {
			body[param] = v
		}
	}
	client, err := c.API()
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := client.Post("/nurserio/orders", body, &out); err != nil {
		return nil, err
	}
	id := str(dig(out, "order", "id"))
	return &cli.Result{Data: out, Summary: fmt.Sprintf(msg.ResOrderCreated, str(dig(out, "order", "order_number"))),
		Crumbs: []cli.Crumb{{Action: msg.CrumbPreparerReserverLeStock, Cmd: "terranova nurserio orders process " + id}}}, nil
}
