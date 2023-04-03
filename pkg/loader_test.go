package owners

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLoader(t *testing.T) {
	fs := afero.NewMemMapFs()

	fs.MkdirAll("a/b", 0755)
	afero.WriteFile(fs, "OWNERS", []byte("root.go @root"), 0644)
	afero.WriteFile(fs, "a/OWNERS", []byte("a.go @a"), 0644)

	loader := newLoaderWithFs("OWNERS", fs)

	// Loads root OWNERS file with empty argument.
	ownersFile, err := loader.Load("")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads root OWNERS file with . argument.
	ownersFile, err = loader.Load(".")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "root.go", Owners: []string{"@root"}},
		}},
	}}, ownersFile)

	// Loads a/OWNERS file.
	ownersFile, err = loader.Load("a")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{Sections: []*Section{
		{Name: defaultSectionName, Approvals: 1, Rules: []*Rule{
			{Pattern: "a.go", Owners: []string{"@a"}},
		}},
	}}, ownersFile)

	// Loads proxy empty file for directory without OWNERS file.
	ownersFile, err = loader.Load("a/b")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)

	// Loads proxy empty file for non existant directory.
	ownersFile, err = loader.Load("a/b/c/d")
	assert.NoError(t, err)
	assert.Equal(t, &OwnersFile{}, ownersFile)
}
