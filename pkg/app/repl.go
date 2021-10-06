package app

import (
	"fmt"
	"path/filepath"
	"time"

	runtime "github.com/aserto-dev/aserto-runtime"
	"github.com/open-policy-agent/opa/repl"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Repl(ref string, maxErrors int) error {
	defer c.Cancel()

	err := c.Pull(ref)
	if err != nil {
		return err
	}

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()
	existingRefParsed, err := c.calculatePolicyRef(ref)
	if err != nil {
		return err
	}

	descriptor, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", ref)
	}

	bundleFile := filepath.Join(c.Configuration.PoliciesRoot(), "blobs", "sha256", descriptor.Digest.Hex())

	opaRuntime, cleanup, err := runtime.NewRuntime(c.Context, c.Logger, &runtime.Config{
		InstanceID: "policy-run",
		LocalBundles: runtime.LocalBundlesConfig{
			Paths: []string{bundleFile},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to setup the OPA runtime")
	}
	defer cleanup()

	err = opaRuntime.PluginsManager.Start(c.Context)
	if err != nil {
		return errors.Wrap(err, "OPA runtime failed to start")
	}

	err = opaRuntime.WaitForPlugins(c.Context, time.Minute*1)
	if err != nil {
		return errors.Wrap(err, "plugins didn't start on time")
	}

	loop := repl.New(opaRuntime.Store, c.Configuration.ReplHistoryFile(), c.UI.Output(), "", maxErrors, fmt.Sprintf("running policy [%s]", ref))
	loop.Loop(c.Context)

	return nil
}

// func (c *PolicyApp) startRunShell(opaRuntime *runtime.Runtime) {
// 	os.Args = []string{}

// 	var shell = grumble.New(&grumble.Config{
// 		Name:        "Policy Interactive Shell",
// 		Description: "Run and debug queries using a loaded policy.",
// 		Prompt:      ">> ",
// 	})

// 	shell.Printf("\n\nPolicy Interactive Shell (you can run 'help' for some pointers)\n\n")

// 	var (
// 		input map[string]interface{}
// 	)

// 	shell.AddCommand(&grumble.Command{
// 		Name: "input",
// 		Help: "Set input data for the query.",
// 		Args: func(a *grumble.Args) {
// 			a.String("input", "input JSON for your query", grumble.Default("{}"))
// 		},
// 		Run: func(s *grumble.Context) error {
// 			err := json.Unmarshal([]byte(s.Args.String("input")), &input)
// 			if err != nil {
// 				c.UI.Problem().WithErr(err).Msg("Invalid JSON")
// 			} else {
// 				c.UI.Normal().Msg("Input set.")
// 			}

// 			return nil
// 		},
// 	})

// 	shell.AddCommand(&grumble.Command{
// 		Name: "query",
// 		Help: "Run a query.",
// 		Args: func(a *grumble.Args) {
// 			a.String("query", "query to run", grumble.Default("x=data"))
// 		},
// 		Run: func(s *grumble.Context) error {
// 			result, err := opaRuntime.Query(
// 				context.Background(),
// 				s.Args.String("query"),
// 				input,
// 				true,
// 				false,
// 				false,
// 				"off",
// 			)

// 			if err != nil {
// 				c.UI.Problem().WithErr(err).Msg("Query failed.")
// 			} else {

// 				out, err := json.MarshalIndent(result.Result, "", "  ")
// 				if err != nil {
// 					c.UI.Problem().WithErr(err).Msg("Can't marshal result JSON.")
// 				}

// 				c.UI.Normal().Compact().Msg(string(out))
// 			}
// 			return nil
// 		},
// 	})

// 	// run shell
// 	err := shell.Run()
// 	if err != nil {
// 		c.UI.Problem().WithErr(err).Do()
// 	}
// }
