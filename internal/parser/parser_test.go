//nolint:goconst
package parser_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/opcr-io/policy/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDomain(t *testing.T) {
	assert := assert.New(t)

	defaultDomains := []string{"", "docker.io", "ghcr.io"}

	tcs := []struct {
		input  string
		output string
	}{
		// no explicit domain specified, inherits from default domain setting
		{"foo/bar:baz", "${DOMAIN}/foo/bar:baz"},
		{"foo/bar", "${DOMAIN}/foo/bar:latest"},
		{"foo", "${DOMAIN}/library/foo:latest"},
		{"default", "${DOMAIN}/library/default:latest"},
		{"default:latest", "${DOMAIN}/library/default:latest"},
		{"default:0.0.1", "${DOMAIN}/library/default:0.0.1"},
		{"library/default", "${DOMAIN}/library/default:latest"},
		{"library/default:latest", "${DOMAIN}/library/default:latest"},
		{"test", "${DOMAIN}/library/test:latest"},
		{"test:latest", "${DOMAIN}/library/test:latest"},
		{"test:0.0.1", "${DOMAIN}/library/test:0.0.1"},
		// fixed docker.io domain anchored tags
		{"docker.io/default", "docker.io/library/default:latest"},
		{"docker.io/default:latest", "docker.io/library/default:latest"},
		{"docker.io/library/default", "docker.io/library/default:latest"},
		{"docker.io/library/default:latest", "docker.io/library/default:latest"},
		{"docker.io/foo/bar:baz", "docker.io/foo/bar:baz"},
		{"docker.io/library/nginx", "docker.io/library/nginx:latest"},
		{"docker.io/library/ubuntu", "docker.io/library/ubuntu:latest"},
		{"docker.io/library/ubuntu:latest", "docker.io/library/ubuntu:latest"},
		{"docker.io/test:0.0.2", "docker.io/library/test:0.0.2"},
		// fixed ghcr.io domain anchored tags
		{"ghcr.io/default", "ghcr.io/default:latest"},
		{"ghcr.io/default:latest", "ghcr.io/default:latest"},
		{"ghcr.io/library/default", "ghcr.io/library/default:latest"},
		{"ghcr.io/library/default:latest", "ghcr.io/library/default:latest"},
		{"ghcr.io/foo/bar:baz", "ghcr.io/foo/bar:baz"},
		{"ghcr.io/org/img", "ghcr.io/org/img:latest"},
		{"ghcr.io/linuxcontainers/alpine", "ghcr.io/linuxcontainers/alpine:latest"},
		{"ghcr.io/linuxfiles/alpine", "ghcr.io/linuxfiles/alpine:latest"},
		// random test cases
		{"alpine", "${DOMAIN}/library/alpine:latest"},
		{"nginx", "${DOMAIN}/library/nginx:latest"},
		{"library/nginx", "${DOMAIN}/library/nginx:latest"},
		{"ubuntu:latest", "${DOMAIN}/library/ubuntu:latest"},
		{"library/ubuntu:latest", "${DOMAIN}/library/ubuntu:latest"},
		{"mycompany.registry:5000/myapp", "mycompany.registry:5000/myapp:latest"},
		{"localhost:5000/foo", "localhost:5000/foo:latest"},
		{"my-registry.local:5000/internal/app:v1.0", "my-registry.local:5000/internal/app:v1.0"},
		{"invalid__domain/image:latest", "${DOMAIN}/invalid__domain/image:latest"},
	}

	for d, defaultDomain := range defaultDomains {
		for i, tc := range tcs {
			output := os.Expand(tc.output, func(key string) string {
				switch key {
				case "DOMAIN":
					if defaultDomain == "" {
						return parser.DefaultCanonicalDomain
					}

					return defaultDomain

				default:
					return ""
				}
			})

			t.Run(fmt.Sprintf("tc:%d%03d", d, i), func(t *testing.T) {
				computedRef, err := parser.CalculateRef(tc.input, defaultDomain)
				if err != nil {
					assert.FailNow(err.Error())
				}

				t.Logf("tc:%03d default:%q input:%q output:%q computed:%q", i, defaultDomain, tc.input, output, computedRef)

				if !strings.EqualFold(output, computedRef) {
					assert.Equal(output, computedRef, "tc:%d - %q", i, tc.input)
				}
			})
		}
	}
}
