package parser

import (
	"github.com/containerd/containerd/reference/docker"
	"github.com/pkg/errors"
)

const (
	DefaultCanonicalDomain = "docker.io"
)

// Calculates the docker reference from string.
func CalculatePolicyRef(userRef, defaultDomain string) (string, error) {
	parsed, err := docker.ParseDockerRef(userRef)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse reference [%s]", userRef)
	}

	familiarized := docker.FamiliarString(parsed)

	domain := docker.Domain(parsed)

	if (familiarized == userRef || familiarized == userRef+":latest") && domain == DefaultCanonicalDomain {
		parsedWithDomain, err := docker.ParseDockerRef(defaultDomain + "/" + userRef)
		if err != nil {
			return "", errors.Wrapf(err, "failed to parse normalized reference [%s]", defaultDomain+"/"+userRef)
		}

		return parsedWithDomain.String(), nil
	}

	return userRef, nil
}
