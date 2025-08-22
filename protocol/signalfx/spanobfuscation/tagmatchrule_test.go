package spanobfuscation

import (
	"testing"

	"github.com/gobwas/glob"
	"github.com/gobwas/glob/match"
	"github.com/signalfx/golib/v3/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRules(t *testing.T) {
	defaultGlob, _ := glob.Compile(`*`)
	serviceGlob, _ := glob.Compile(`\^\\some*service\$`)
	opGlob, _ := glob.Compile(`operation\.*`)

	cases := []struct {
		desc        string
		config      []*TagMatchRuleConfig
		outputRules []*rule
	}{
		{
			desc:        "empty service and empty operation",
			config:      []*TagMatchRuleConfig{{Tags: []string{"test-tag"}}},
			outputRules: []*rule{{service: defaultGlob, operation: defaultGlob, tags: []string{"test-tag"}}},
		},
		{
			desc:        "service regex and empty operation",
			config:      []*TagMatchRuleConfig{{Service: pointer.String(`^\some*service$`), Tags: []string{"test-tag"}}},
			outputRules: []*rule{{service: serviceGlob, operation: defaultGlob, tags: []string{"test-tag"}}},
		},
		{
			desc:        "empty service and operation regex",
			config:      []*TagMatchRuleConfig{{Operation: pointer.String(`operation.*`), Tags: []string{"test-tag"}}},
			outputRules: []*rule{{service: defaultGlob, operation: opGlob, tags: []string{"test-tag"}}},
		},
		{
			desc:        "service regex and operation regex",
			config:      []*TagMatchRuleConfig{{Service: pointer.String(`^\some*service$`), Operation: pointer.String(`operation.*`), Tags: []string{"test-tag"}}},
			outputRules: []*rule{{service: serviceGlob, operation: opGlob, tags: []string{"test-tag"}}},
		},
		{
			desc:        "multiple tags",
			config:      []*TagMatchRuleConfig{{Service: pointer.String(`^\some*service$`), Operation: pointer.String(`operation.*`), Tags: []string{"test-tag", "another-tag"}}},
			outputRules: []*rule{{service: serviceGlob, operation: opGlob, tags: []string{"test-tag", "another-tag"}}},
		},
		{
			desc: "multiple rules",
			config: []*TagMatchRuleConfig{
				{Tags: []string{"test-tag"}},
				{Service: pointer.String(`^\some*service$`), Tags: []string{"test-tag"}},
				{Operation: pointer.String(`operation.*`), Tags: []string{"test-tag"}},
				{Service: pointer.String(`^\some*service$`), Operation: pointer.String(`operation.*`), Tags: []string{"test-tag"}},
			},
			outputRules: []*rule{
				{service: defaultGlob, operation: defaultGlob, tags: []string{"test-tag"}},
				{service: serviceGlob, operation: defaultGlob, tags: []string{"test-tag"}},
				{service: defaultGlob, operation: opGlob, tags: []string{"test-tag"}},
				{service: serviceGlob, operation: opGlob, tags: []string{"test-tag"}},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			r, err := getRules(tc.config)
			require.NoError(t, err)
			require.NotNil(t, r)
			for i := 0; i < len(r); i++ {
				assert.Equal(t, tc.outputRules[i].service.(match.Matcher).String(), r[i].service.(match.Matcher).String())
				assert.Equal(t, tc.outputRules[i].operation.(match.Matcher).String(), r[i].operation.(match.Matcher).String())
				for idx, tag := range r[i].tags {
					assert.Equal(t, tc.outputRules[i].tags[idx], tag)
				}
			}
		})
	}
}
