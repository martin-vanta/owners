package owners

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSectionHeader(t *testing.T) {
	tests := []struct {
		line     string
		expected *Section
	}{
		{line: "", expected: nil},
		{line: "[]", expected: nil},
		{line: "^[]", expected: nil},
		{
			line:     "[name]",
			expected: &Section{Name: "name", Approvals: 1},
		},
		{
			line:     "^[name]",
			expected: &Section{Name: "name", Optional: true, Approvals: 1},
		},
		{
			line:     "[name][10]",
			expected: &Section{Name: "name", Approvals: 10},
		},
		{
			line:     "[name][1]",
			expected: &Section{Name: "name", Approvals: 1},
		},
		{
			line:     "[name][0]",
			expected: &Section{Name: "name", Approvals: 1},
		},
		{
			line:     "[name][-1]",
			expected: &Section{Name: "name", Approvals: 1},
		},
		{
			line:     "[name] @user1",
			expected: &Section{Name: "name", DefaultOwners: []string{"@user1"}, Approvals: 1},
		},
		{
			line:     "[name] @user1 @user2",
			expected: &Section{Name: "name", DefaultOwners: []string{"@user1", "@user2"}, Approvals: 1},
		},
		{
			line:     "[name] @org/user1",
			expected: &Section{Name: "name", DefaultOwners: []string{"@org/user1"}, Approvals: 1},
		},
		{
			line:     "^[name][2] @user1 @org/user2",
			expected: &Section{Name: "name", Optional: true, DefaultOwners: []string{"@user1", "@org/user2"}, Approvals: 2},
		},
	}
	for _, test := range tests {
		got := parseSectionHeader(test.line)
		assert.Equal(t, test.expected, got, "line: '%s'", test.line)
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		contents string
		expected *RuleFile
	}{
		{contents: "", expected: &RuleFile{Sections: []*Section{{Name: defaultSectionName, Approvals: 1}}}},
		{contents: "  ", expected: &RuleFile{Sections: []*Section{{Name: defaultSectionName, Approvals: 1}}}},
		{contents: "  # foo", expected: &RuleFile{Sections: []*Section{{Name: defaultSectionName, Approvals: 1}}}},
		{
			contents: `
				foo.go @user1
				bar.ts @user2
			`,
			expected: &RuleFile{Sections: []*Section{
				{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
					{Pattern: "foo.go", Owners: []string{"@user1"}},
					{Pattern: "bar.ts", Owners: []string{"@user2"}},
				}},
			}},
		},
		{
			contents: `
				[go]
				foo.go @user1
				^[docs] @docs
				readme.md
			`,
			expected: &RuleFile{Sections: []*Section{
				{Name: defaultSectionName, Approvals: 1},
				{Name: "go", Approvals: 1, Rules: []*Rule{
					{Pattern: "foo.go", Owners: []string{"@user1"}},
				}},
				{Name: "docs", Optional: true, Approvals: 1, DefaultOwners: []string{"@docs"}, Rules: []*Rule{
					{Pattern: "readme.md", Owners: []string{}},
				}},
			}},
		},
	}
	for _, test := range tests {
		got, err := ParseFile(bytes.NewBufferString(test.contents))
		assert.NoError(t, err)
		assert.Equal(t, test.expected, got, "contents:\n%s", test.contents)
	}
}
