<img src="assets/policy.png" alt="policy">

# policy - the CLI for managing authorization policies

The policy CLI is a tool for building, versioning and publishing your authorization policies.
It uses OCI standards to manage artifacts, and the [Open Policy Agent (OPA)](https://github.com/open-policy-agent/opa) to compile and run.

---

[![Go Report Card](https://goreportcard.com/badge/github.com/opcr-io/policy?)](https://goreportcard.com/report/github.com/opcr-io/policy)
[![ci](https://github.com/opcr-io/policy/actions/workflows/ci.yaml/badge.svg)](https://github.com/opcr-io/policy/actions/workflows/ci.yaml)
[![codebeat badge](https://codebeat.co/badges/8e9c8690-9890-46d4-accc-17e5ac24cd88)](https://codebeat.co/projects/github-com-opcr-io-policy-main)
![GitHub all releases](https://img.shields.io/github/downloads/opcr-io/policy/total)
![Apache 2.0](https://img.shields.io/github/license/opcr-io/policy)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/opcr-io/policy)
[<img src="https://img.shields.io/badge/slack-@asertocommunity-yellow.svg?logo=slack">](https://asertocommunity.slack.com/)
[<img src="https://img.shields.io/badge/docs-%F0%9F%95%B6-blue">](https://www.openpolicycontainers.com/docs/intro)
[![CodeQL](https://github.com/opcr-io/policy/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/opcr-io/policy/actions/workflows/codeql-analysis.yml)
[![OpenSSF Best Practices](https://bestpractices.coreinfrastructure.org/projects/6859/badge)](https://bestpractices.coreinfrastructure.org/projects/6859)
---

## Documentation

Please refer to our [documentation](https://openpolicycontainers.com) site for installation, usage, customization and tips.

## Slack Channel

Wanna discuss features or show your support for this tool?

* Channel: [Slack](https://asertocommunity.slack.com/)
* Invite: [Invite Link](https://asertocommunity.slack.com/join/shared_invite/zt-p06gin84-xNswWpTGyPDPxCz0LMux3g#/shared-invite/email)

---

## Installation

`policy` is available on Linux, macOS and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tarballs in the [release](https://github.com/opcr-io/policy/releases) page.

* Via Homebrew for macOS or LinuxBrew for Linux

   ```shell
  brew tap opcr-io/tap && brew install opcr-io/tap/policy
   ```

* Via the nix package manager on nixOS, other linux distros, and macOS

   At the moment the package is only available in the `unstable` channel. Below are some examples using nix to install `policy` via the shell, NixOS configuration, and home-manager configuration.

   Shell:
   ```shell
   nix-env --install -A nixpkgs.opcr-policy
   ```

   NixOS:
   ```nix
     # your other config ...
     environment.systemPackages = with pkgs; [
       # your other packages ...
       opcr-policy
     ];

   ```

   home-manager:
   ```nix
     # your other config ...
     home.packages = with pkgs; [
       # your other packages ...
       opcr-policy
     ];
   ```

* Via a GO install

  ```shell
  go install github.com/opcr-io/policy/cmd/policy@latest
  ```

---

## Building From Source

 `policy` is currently using go v1.16 or above. In order to build `policy` from source you must:

 1. Install [mage](https://magefile.org/)
 2. Clone the repo
 3. Build and run the executable

      ```shell
      mage build && ./dist/build_linux_amd64/policy
      ```

---

## Running with Docker

### Running the official Docker image

  You can run as a Docker container:

  ```shell
  docker run -it --rm ghcr.io/opcr-io/policy:latest --help
  ```


---

## The Command Line

```shell
$ policy --help
Usage: policy <command>

Flags:
  -h, --help             Show context-sensitive help.
  -c, --config="/Users/ogazitt/.policy/config.yaml"
                         Path to the policy CLI config file.
      --debug            Enable debug mode.
  -v, --verbosity=INT    Use to increase output verbosity.
  -k, --insecure         Do not verify TLS connections.

Commands:
  build <path> ...
    Build policies.

  images
    List policy images.

  push <policy> ...
    Push policies to a registry.

  pull <policy> ...
    Pull policies from a registry.

  login --server=STRING --username=STRING
    Login to a registry.

  logout
    Logout from a registry.

  save <policy>
    Save a policy to a local bundle tarball.

  tag <policy> <tag>
    Create a new tag for an existing policy.

  rm <policy> ...
    Removes a policy from the local registry.

  inspect <policy>
    Displays information about a policy.

  repl <policy>
    Sets you up with a shell for running queries using an OPA instance with a policy loaded.

  remote set-public <policy> [<public>]
    Make a policy public or private.

  remote images
    Synonym for 'policy images --remote'.

  init [<path>]
    (Deprecated) Initialize policy repo

  templates apply <template>
    Create or update a policy or related artifacts from a template.

  templates list
    List all available templates.

  version
    Prints version information.

Run "policy <command> --help" for more information on a command.
```

## Logs

Logs are printed to `stderr`. You can increase detail using the verbosity flag (e.g. `-vvv`).

## Demo Videos/Recordings

![demo](./assets/demo-policy.gif)

---

## Known Issues

This is still work in progress! If something is broken or there's a feature
that you want, please file an issue and if so inclined submit a PR!

---

## Credits

The policy CLI uses a lot of great and amazing open source projects and libraries.
A big thank you to all of them!

---

## Contributions Guideline

* File an issue first prior to submitting a PR!
* Ensure all exported items are properly commented
* If applicable, submit a test suite against your PR

## Reporting Vulnerabilities

Please send an email to one of the [maintainers](MAINTAINERS.md). We commit to addressing vulnerabilities promptly.

