<img src="assets/policy.png" alt="policy">

# policy - the CLI for managing authorization policies

The policy CLI is a tool for building, versioning and publishing your authorization policies.
It uses OCI standards to manage artifacts, and the [Open Policy Agent (OPA)](https://github.com/open-policy-agent/opa) to compile and run.

---

[![Go Report Card](https://goreportcard.com/badge/github.com/opcr-io/policy?)](https://goreportcard.com/report/github.com/opcr-io/policy)
TODO: add more badges once the repo is open

---

## Documentation

Please refer to our [documentation](https://openpolicyregistry.io) site for installation, usage, customization and tips.

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
   brew install policy
   ```

* Via a GO install

  ```shell
  # NOTE: The dev version will be in effect!
  go get -u github.com/opcr-io/policy
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
  -c, --config="/home/toaster/.config/policy/config.yaml"
                         Path to the policy CLI config file.
      --debug            Enable debug mode.
  -v, --verbosity=INT    Use to increase output verbosity.

Commands:
  build <path> ...
    Build policies.

  list
    List policies.

  push <policy> ...
    Push policies to a registry.

  pull <policy> ...
    Pull policies from a registry.

  login
    Login to a registry.

  save <policy>
    Save a policy to a local bundle tarball.

  tag <policy> <tag>
    Create a new tag for an existing policy.

  rm <policy> ...
    Removes a policy from the local registry.

  run <policy>
    Sets you up with a shell for running queries using an OPA instance with a policy loaded.

  version
    Prints version information.

Run "policy <command> --help" for more information on a command.
```

## Logs

Logs are printed to `stderr`. You can increase detail using the verbosity flag (e.g. `-vvv`).

## Demo Videos/Recordings

TODO: record a demo

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
