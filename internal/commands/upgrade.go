package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/semisto-org/terranova-cli/internal/cli"
)

const releasesLatest = "https://api.github.com/repos/semisto-org/terranova-cli/releases/latest"

func init() {
	cli.Register(&cli.Command{
		Name: "upgrade", Group: "Auth & Config",
		Summary: "Met le binaire à jour en place depuis la dernière release (ISC-425).",
		Run:     runUpgrade,
	})
}

func runUpgrade(c *cli.Ctx, args []string) (*cli.Result, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Get(releasesLatest)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("aucune release publiée")
	}
	if release.TagName == c.Version || "v"+c.Version == release.TagName {
		return &cli.Result{Data: map[string]string{"version": c.Version},
			Summary: "Déjà à jour (" + c.Version + ")."}, nil
	}
	want := fmt.Sprintf("terranova_%s_%s", runtime.GOOS, runtime.GOARCH)
	var url string
	for _, a := range release.Assets {
		if a.Name == want {
			url = a.URL
		}
	}
	if url == "" {
		return nil, fmt.Errorf("pas de binaire %s dans la release %s", want, release.TagName)
	}
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	tmp := self + ".new"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return nil, err
	}
	dl, err := client.Get(url)
	if err != nil {
		out.Close()
		return nil, err
	}
	defer dl.Body.Close()
	if _, err := io.Copy(out, dl.Body); err != nil {
		out.Close()
		return nil, err
	}
	out.Close()
	if err := os.Rename(tmp, self); err != nil {
		return nil, err
	}
	return &cli.Result{Data: map[string]string{"from": c.Version, "to": release.TagName},
		Summary: "Mis à jour : " + c.Version + " → " + release.TagName + "."}, nil
}
