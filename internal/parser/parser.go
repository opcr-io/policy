package parser

import (
	"strings"

	"github.com/distribution/reference"
)

const (
	DefaultCanonicalDomain = "docker.io"
)

func CalculateRef(userRef, defaultDomain string) (string, error) {
	ref, err := CalculateNamedRef(userRef, defaultDomain)
	if err != nil {
		return "", err
	}

	return ref.String(), nil
}

func CalculateNamedRef(userRef, defaultDomain string) (reference.Named, error) {
	if defaultDomain == "" {
		defaultDomain = DefaultCanonicalDomain
	}

	dockerRef, err := reference.ParseDockerRef(userRef)
	if err != nil {
		return nil, err
	}

	incomingDomain := reference.Domain(dockerRef)

	hasDomain := strings.HasPrefix(userRef, incomingDomain)

	if defaultDomain != DefaultCanonicalDomain && incomingDomain != defaultDomain && !hasDomain {
		tmpRef := defaultDomain + strings.TrimPrefix(dockerRef.String(), DefaultCanonicalDomain)

		updatedRef, err := reference.ParseDockerRef(tmpRef)
		if err != nil {
			return nil, err
		}

		return updatedRef, nil
	}

	return dockerRef, nil
}
