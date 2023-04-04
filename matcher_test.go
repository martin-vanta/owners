package owners

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMatcherMatch(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := fs.MkdirAll("a/b", 0755)
	assert.NoError(t, err)
	err = afero.WriteFile(fs, "OWNERS", []byte(`
		[required]
		root.go @root
		a/a.go @root_overridden
		/root_slash.go @root_slash
		.//./root_slash_unnormalized.go @root_slash_unnormalized

		a/**/*.s @doublestar
		b/*.s @singlestar
		**/doublestar_prefix.s @doublestar_prefix

		^[optional]
		root_optional.go @root_optional
		`), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(fs, "a/OWNERS", []byte(`
		[required]
		a.go @a_overridden
		a.go @a
		a_both.go @a
		/a_slash.go @a_slash

		^[optional]
		a_optional.go @a_optional
		a_both.go @a_optional
		`), 0644)
	assert.NoError(t, err)

	matcher := newMatcherWithFs("OWNERS", fs)

	tests := []struct {
		filePath string
		expected []MatchOwner
	}{
		{filePath: "", expected: nil},
		{filePath: ".", expected: nil},
		{filePath: "does_not_exist.go", expected: nil},
		{filePath: "a/does_not_exist.go", expected: nil},
		{filePath: "a/b/c/d/does_not_exist.go", expected: nil},

		{filePath: "root.go", expected: []MatchOwner{{Owner: "@root"}}},
		{filePath: "root_optional.go", expected: []MatchOwner{{Owner: "@root_optional", Optional: true}}},
		{filePath: "root_slash.go", expected: []MatchOwner{{Owner: "@root_slash"}}},
		{filePath: "root_slash_unnormalized.go", expected: []MatchOwner{{Owner: "@root_slash_unnormalized"}}},

		{filePath: "a/a.go", expected: []MatchOwner{{Owner: "@a"}}},
		{filePath: "a/a_optional.go", expected: []MatchOwner{{Owner: "@a_optional", Optional: true}}},
		{filePath: "a/a_both.go", expected: []MatchOwner{{Owner: "@a"}, {Owner: "@a_optional", Optional: true}}},
		{filePath: "a/a_slash.go", expected: []MatchOwner{{Owner: "@a_slash"}}},

		{filePath: "doublestar_prefix.s", expected: []MatchOwner{{Owner: "@doublestar_prefix"}}},
		{filePath: "a/doublestar_prefix.s", expected: []MatchOwner{{Owner: "@doublestar_prefix"}}},
		{filePath: "a/b/c/d/doublestar_prefix.s", expected: []MatchOwner{{Owner: "@doublestar_prefix"}}},

		{filePath: "a/doublestar.s", expected: []MatchOwner{{Owner: "@doublestar"}}},
		{filePath: "a/b/c/d/doublestar.s", expected: []MatchOwner{{Owner: "@doublestar"}}},

		{filePath: "b/singlestar.s", expected: []MatchOwner{{Owner: "@singlestar"}}},
		{filePath: "b/c/d/singlestar.s", expected: nil},
	}
	for _, test := range tests {
		got, err := matcher.Match(test.filePath)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, got, "file: %s", test.filePath)
	}
}

func TestMatcherLoad(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := fs.MkdirAll("a/b", 0755)
	assert.NoError(t, err)
	err = afero.WriteFile(fs, "OWNERS", []byte("root.go @root"), 0644)
	assert.NoError(t, err)
	err = afero.WriteFile(fs, "a/OWNERS", []byte("a.go @a"), 0644)
	assert.NoError(t, err)

	matcher := newMatcherWithFs("OWNERS", fs)

	// Loads root OWNERS file with empty argument.
	ownersFile, err := matcher.Load("")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads root OWNERS file with . argument.
	ownersFile, err = matcher.Load(".")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads a/OWNERS file.
	ownersFile, err = matcher.Load("a")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "a.go", Owners: []string{"@a"}},
		}},
	}}, ownersFile)

	// Loads proxy empty file for directory without OWNERS file.
	ownersFile, err = matcher.Load("a/b")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)

	// Loads proxy empty file for non existant directory.
	ownersFile, err = matcher.Load("a/b/c/d")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)
}
