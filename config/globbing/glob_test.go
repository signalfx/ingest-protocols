package globbing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeMetaCharacters(t *testing.T) {
	cases := []struct {
		desc    string
		pattern string
		match   []string
		noMatch []string
	}{
		{
			desc:    `^`,
			pattern: `^test*service`,
			match:   []string{`^test-service`, `^test^service`},
			noMatch: []string{`test-service`, `testservice`},
		},
		{
			desc:    `?`,
			pattern: `test?service*`,
			match:   []string{`test?service-one`},
			noMatch: []string{`test-service`, `test.service`, `testaservice`, `test.service.prod`},
		},
		{
			desc:    `\`,
			pattern: `test\*service`,
			match:   []string{`test\this\service`, `test\service`, `test\\service`},
			noMatch: []string{`test-service`, `test.service`, `testservice`},
		},
		{
			desc:    `{}`,
			pattern: `service{2}`,
			match:   []string{`service{2}`},
			noMatch: []string{`servicee`, `serviceeee`},
		},
		{
			desc:    `[]`,
			pattern: `service*[a-z]`,
			match:   []string{`service[a-z]`, `service/handle/[a-z]`},
			noMatch: []string{`servicea`, `servicee`},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			g := GetGlob(c.pattern)
			for _, m := range c.match {
				assert.True(t, g.Match(m), "pattern %q should match %q", c.pattern, m)
			}
			for _, n := range c.noMatch {
				assert.False(t, g.Match(n), "pattern %q should not match %q", c.pattern, n)
			}
		})
	}
}
