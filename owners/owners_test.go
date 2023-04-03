package owners

import (
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
