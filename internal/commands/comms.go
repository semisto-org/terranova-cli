package commands

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/semisto-org/terranova-cli/internal/cli"
	"github.com/semisto-org/terranova-cli/internal/msg"
)

// Les surfaces nées côté app dans la même nuit que ce fichier (ISC-401/407/408) :
// recherche globale, lignes de chat, notifications — plus `my`, les vues du matin.
func init() {
	cli.Register(&cli.Command{
		Name: "search", Group: msg.GroupSearchBrowse,
		Summary: msg.HelpSearch,
		ArgSpec: "<termes…>", MinArgs: 1,
		Flags: []cli.Flag{
			{Name: "type", Arg: "Type", Help: msg.FlagSearchType},
		},
		APIOps: []string{"GET /search"},
		Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
			typ, args := cli.FlagValue(args, "type")
			q := strings.Join(args, " ")
			client, err := c.API()
			if err != nil {
				return nil, err
			}
			path := "/search?q=" + url.QueryEscape(q)
			if typ != "" {
				path += "&type=" + url.QueryEscape(typ)
			}
			if c.Flags.Project != "" {
				path += "&project_id=" + url.QueryEscape(c.Flags.Project)
			}
			var out struct {
				Recordings []map[string]any `json:"recordings"`
				Projects   []map[string]any `json:"projects"`
			}
			if err := client.Get(path, &out); err != nil {
				return nil, err
			}
			rows := [][]string{}
			for _, p := range out.Projects {
				rows = append(rows, []string{str(p["id"]), msg.LabelProjet, str(p["name"]), ""})
			}
			for _, r := range out.Recordings {
				rows = append(rows, []string{str(r["id"]), str(r["recordable_type"]), str(r["title"]), str(r["bucket_id"])})
			}
			return &cli.Result{Data: out, Headers: msg.HeadersTypedList, Rows: rows,
				Summary: fmt.Sprintf(msg.ResSearchCount, len(out.Recordings), len(out.Projects)),
				Crumbs:  []cli.Crumb{{Action: msg.CrumbOuvrirUnResultat, Cmd: "terranova recordings show <id>"}}}, nil
		},
	})

	cli.Register(&cli.Command{
		Name: "notifications", Group: msg.GroupCommunication,
		Summary: msg.HelpNotifications,
		Sub: []*cli.Command{
			{
				Name: "list", Summary: msg.HelpNotificationsList,
				Flags:  []cli.Flag{{Name: "unread", Help: msg.FlagNotificationsListUnread}},
				APIOps: []string{"GET /notifications"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					unread, _ := cli.FlagBool(args, "unread")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/notifications"
					if unread {
						path += "?unread=1"
					}
					var out struct {
						Notifications []map[string]any `json:"notifications"`
					}
					if err := client.Get(path, &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, n := range out.Notifications {
						read := "•"
						if b, ok := n["read"].(bool); ok && b {
							read = ""
						}
						rows = append(rows, []string{read, str(n["id"]), str(n["kind"]), str(n["actor"]), str(n["title"])})
					}
					return &cli.Result{Data: out.Notifications, Headers: msg.HeadersNotifications, Rows: rows,
						Summary: fmt.Sprintf(msg.ResNotificationCount, len(out.Notifications)),
						Crumbs: []cli.Crumb{
							{Action: msg.CrumbMarquerLu, Cmd: "terranova notifications read <id>"},
							{Action: msg.CrumbToutMarquerLu, Cmd: "terranova notifications read-all"},
						}}, nil
				},
			},
			{
				Name: "read", Summary: msg.HelpNotificationsRead, ArgSpec: "<id>", MinArgs: 1,
				Flags:  []cli.Flag{{Name: "unread", Help: msg.FlagNotificationsReadUnread}},
				APIOps: []string{"POST /notifications/{id}/read"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					unread, rest := cli.FlagBool(args, "unread")
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					path := "/notifications/" + rest[0] + "/read"
					var body any
					if unread {
						body = map[string]any{"read": "0"}
					}
					var out map[string]any
					if err := client.Post(path, body, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: msg.ResFait}, nil
				},
			},
			{
				Name: "read-all", Summary: msg.HelpNotificationsReadAll,
				APIOps: []string{"POST /notifications/read_all"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out map[string]any
					if err := client.Post("/notifications/read_all", nil, &out); err != nil {
						return nil, err
					}
					return &cli.Result{Data: out, Summary: fmt.Sprintf(msg.ResMarkedRead, str(out["marked_read"]))}, nil
				},
			},
		},
	})

	// `my` — les vues du matin (ISC-410, la part que l'API offre déjà).
	cli.Register(&cli.Command{
		Name: "my", Group: msg.GroupSchedulingTime,
		Summary: msg.HelpMy,
		Sub: []*cli.Command{
			{
				Name: "todos", Summary: msg.HelpMyTodos,
				APIOps: []string{"GET /me", "GET /recordings"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var me struct {
						Me struct {
							ID any `json:"id"`
						} `json:"me"`
					}
					if err := client.Get("/me", &me); err != nil {
						return nil, err
					}
					var out struct {
						Recordings []map[string]any `json:"recordings"`
					}
					if err := client.Get("/recordings?type=Todo&assigned_to="+str(me.Me.ID), &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, r := range out.Recordings {
						due := str(dig(r, "todo", "due_on"))
						rows = append(rows, []string{str(r["id"]), str(r["title"]), due})
					}
					return &cli.Result{Data: out.Recordings, Headers: msg.HeadersMyTodos, Rows: rows,
						Summary: fmt.Sprintf(msg.ResAssignedTodoCount, len(out.Recordings)),
						Crumbs:  []cli.Crumb{{Action: msg.CrumbCompleter, Cmd: "terranova todos edit <id> --completed true"}}}, nil
				},
			},
			{
				Name: "timesheets", Summary: msg.HelpMyTimesheets,
				APIOps: []string{"GET /my/timesheets"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Timesheets []map[string]any `json:"timesheets"`
						TotalHours float64          `json:"total_hours"`
					}
					if err := client.Get("/my/timesheets", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, t := range out.Timesheets {
						rows = append(rows, []string{str(t["worked_on"]), str(t["project"]), str(t["hours"]), str(t["description"])})
					}
					return &cli.Result{Data: out, Headers: msg.HeadersMyTimesheets, Rows: rows,
						Summary: fmt.Sprintf(msg.ResHoursTotal, out.TotalHours)}, nil
				},
			},
			{
				Name: "bookmarks", Summary: msg.HelpMyBookmarks,
				APIOps: []string{"GET /my/bookmarks"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					client, err := c.API()
					if err != nil {
						return nil, err
					}
					var out struct {
						Bookmarks []map[string]any `json:"bookmarks"`
					}
					if err := client.Get("/my/bookmarks", &out); err != nil {
						return nil, err
					}
					rows := [][]string{}
					for _, b := range out.Bookmarks {
						rows = append(rows, []string{str(b["id"]), str(b["type"]), str(b["title"]), str(b["project"])})
					}
					return &cli.Result{Data: out.Bookmarks, Headers: msg.HeadersTypedList, Rows: rows,
						Summary: fmt.Sprintf(msg.ResBookmarkCount, len(out.Bookmarks))}, nil
				},
			},
			{
				Name: "notifications", Summary: msg.HelpMyNotifications,
				APIOps: []string{"GET /notifications"},
				Run: func(c *cli.Ctx, args []string) (*cli.Result, error) {
					sub, _ := cli.Find([]string{"notifications", "list"})
					return sub.Run(c, []string{"--unread"})
				},
			},
		},
	})
}
