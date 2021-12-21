package main

import "github.com/pkg/errors"

type BuildCmd struct {
	Tag         string            `name:"tag" short:"t" help:"Name and optionally a tag in the 'name:tag' format"`
	Path        []string          `arg:"" name:"path" help:"Path to the policy sources." type:"string"`
	Annotations map[string]string `name:"annotations" short:"a" help:"Annotations to apply to the policy." type:"string:string"`

	RunConfigFile      string   `name:"build-config-file" help:"Set path of configuration file."`
	Target             string   `name:"target" default:"rego" help:"Set the output bundle target type."`
	OptimizationLevel  int      `name:"optimize" short:"O" default:"0" help:"Set optimization level."`
	Entrypoints        []string `name:"entrypoint" short:"e" help:"Set slash separated entrypoint path."`
	Revision           string   `name:"revision" short:"r" help:"Set output bundle revision."`
	Ignore             []string `name:"ignore" help:"Set file and directory names to ignore during loading (e.g., '.*' excludes hidden files)."`
	Capabilities       string   `name:"capabilities" help:"Set capabilities.json file path."`
	VerificationKey    string   `name:"verification-key" help:"Set the secret (HMAC) or path of the PEM file containing the public key (RSA and ECDSA)."`
	VerificationKeyID  string   `name:"verification-key-id" default:"default" help:"Name assigned to the verification key used for bundle verification."`
	Algorithm          string   `name:"signing-alg" default:"RS256" help:"Name of the signing algorithm."`
	Scope              string   `name:"scope" help:"Scope to use for bundle signature verification."`
	ExcludeVerifyFiles []string `name:"exclude-files-verify" help:"Set file names to exclude during bundle verification."`
	SigningKey         string   `name:"signing-key" help:"Set the secret (HMAC) or path of the PEM file containing the private key (RSA and ECDSA)."`
	ClaimsFile         string   `name:"claims-file" help:"Set path of JSON file containing optional claims (see: https://openpolicyagent.org/docs/latest/management/#signature-format)."`
}

func (c *BuildCmd) Run(g *Globals) error {
	err := g.App.Build(
		c.Tag,
		c.Path,
		c.Annotations,
		c.RunConfigFile,
		c.Target,
		c.OptimizationLevel,
		c.Entrypoints,
		c.Revision,
		c.Ignore,
		c.Capabilities,
		c.VerificationKey,
		c.VerificationKeyID,
		c.Algorithm,
		c.Scope,
		c.ExcludeVerifyFiles,
		c.SigningKey,
		c.ClaimsFile,
	)
	if err != nil {
		return errors.Wrap(err, "Build failed.")
	}

	<-g.App.Context.Done()

	return nil
}
