package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/opcr-io/policy/pkg/policytemplates"
	"github.com/pkg/errors"
)

type InitCmd struct {
	RootPath  string `arg:"" name:"path" required:"" help:"project root path (defaults to current directory)" default:"."`
	User      string `name:"user" short:"u" help:"user name"`
	Server    string `name:"server" short:"s" help:"registry service name"`
	Repo      string `name:"repo" short:"r" help:"repository (org/repo)"`
	TokenName string `name:"token" short:"t" help:"Github Actions secret token name"`
	SCC       string `name:"scc" help:"source code provider"`
	Overwrite bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
}

const (
	templateServer = "opcr.io"
	ciTemplateOrg  = "ci-templates"
	ciTemplateTag  = "latest"
)

func (c *InitCmd) Run(g *Globals) error {
	if c.Server == "" {
		respServer := ""
		defServer := getDefaultServer(g)

		g.App.UI.Normal().Compact().WithAskString(
			fmt.Sprintf("server: (%s)", defServer), &respServer,
		).Do()
		c.Server = iff(respServer != "", respServer, defServer)
	}

	if c.SCC == "" {
		scc, err := getSupportedCIs(g)
		if err != nil {
			return err
		}
		c.SCC = scc
	}

	if c.User == "" {
		respUser := ""
		defUser := getDefaultUser(g, c.Server)

		g.App.UI.Normal().Compact().WithAskString(
			fmt.Sprintf("user  : (%s)", defUser), &respUser,
		).Do()
		c.User = iff(respUser != "", respUser, defUser)
	}

	if c.TokenName == "" {
		respTokenName := ""
		defTokenName := getDefaultTokenName(g, c.Server)

		g.App.UI.Normal().Compact().WithAskString(
			fmt.Sprintf("github secret name: (%s)", defTokenName), &respTokenName,
		).Do()
		c.TokenName = iff(respTokenName != "", respTokenName, defTokenName)
	}

	if c.Repo == "" {
		respRepo := ""
		defRepo := ""
		g.App.UI.Normal().Compact().WithAskString(
			fmt.Sprintf("repo  : (%s)", defRepo), &respRepo,
		).Do()
		c.Repo = iff(respRepo != "", respRepo, defRepo)
	}

	err := g.App.Init(c.RootPath, c.User, c.Server, c.Repo, c.SCC, c.TokenName, c.Overwrite)
	if err != nil {
		return errors.Wrap(err, "Init failed.")
	}

	<-g.App.Context.Done()

	return nil
}

func iff(f bool, s1, s2 string) string {
	if f {
		return s1
	}
	return s2
}

func getDefaultServer(g *Globals) string {
	if len(g.App.Configuration.Servers) == 0 {
		if g.App.Configuration.DefaultDomain != "" {
			return g.App.Configuration.DefaultDomain
		}
		return ""
	}

	servers := []string{}
	for name := range g.App.Configuration.Servers {
		servers = append(servers, name)
	}

	return buildTable("server", servers)
}

func getSupportedCIs(g *Globals) (string, error) {
	ociTemplates := policytemplates.NewOCI(g.App.Context,
		g.App.Logger,
		g.App.TransportWithTrustedCAs(), policytemplates.Config{
			Server:     templateServer,
			PolicyRoot: g.App.Configuration.PoliciesRoot(),
		})
	repos, err := ociTemplates.ListRepos(ciTemplateOrg, ciTemplateTag)
	if err != nil {
		return "", errors.Wrap(err, "failed to list ci-templates")
	}

	return buildTable("source control provider", repos), nil
}

func buildTable(name string, items []string) string {
	sort.Strings(items)

	allowedValues := make([]int, len(items))
	table := g.App.UI.Normal().WithTable("#", name)
	for i, item := range items {
		table.WithTableRow(strconv.Itoa(i+1), item)
		allowedValues[i] = i + 1
	}

	table.Do()
	var response int64
	g.App.UI.Normal().Compact().WithAskInt(fmt.Sprintf("Select %s#", name), &response, allowedValues...).Do()

	return items[response-1]
}

func getDefaultUser(g *Globals, server string) string {
	if s, ok := g.App.Configuration.Servers[server]; ok {
		return s.Username
	}
	return ""
}

func getDefaultTokenName(g *Globals, server string) string {
	if token, ok := g.App.Configuration.TokenDefaults[server]; ok {
		return token
	}

	if token, ok := g.App.Configuration.TokenDefaults[g.App.Configuration.DefaultDomain]; ok {
		return token
	}

	return ""
}
