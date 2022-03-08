package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

type InitCmd struct {
	RootPath  string `arg:"" name:"path" required:"" help:"project root path (defaults to current directory)" default:"."`
	User      string `name:"user" short:"u" help:"user name"`
	Server    string `name:"server" short:"s" help:"registry service name"`
	Repo      string `name:"repo" short:"r" help:"repository (org/repo)"`
	TokenName string `name:"token" short:"t" help:"Github Actions secret token name"`
	SCC       string `name:"scc" help:"source code provider" required:"" enum:"github, gitlab" default:"github"`
	Overwrite bool   `name:"overwrite" help:"overwrite existing files" default:"false"`
	NoSrc     bool   `name:"no-src" help:"do not write src directory" default:"false"`
}

func (c *InitCmd) Run(g *Globals) error {
	if c.Server == "" {
		respServer := ""
		defServer := getDefaultServer(g)

		g.App.UI.Normal().Compact().WithAskString(
			fmt.Sprintf("server: (%s)", defServer), &respServer,
		).Do()
		c.Server = iff(respServer != "", respServer, defServer)
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

	err := g.App.Init(c.RootPath, c.User, c.Server, c.Repo, c.SCC, c.TokenName, c.Overwrite, c.NoSrc)
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

	sort.Strings(servers)

	allowedValues := make([]int, len(servers))
	table := g.App.UI.Normal().WithTable("#", "Server")
	for i, server := range servers {
		table.WithTableRow(strconv.Itoa(i+1), server)
		allowedValues[i] = i + 1
	}

	table.Do()

	var response int64
	g.App.UI.Normal().Compact().WithAskInt("Select server#", &response, allowedValues...).Do()

	return servers[response-1]
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
