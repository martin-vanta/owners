package owners

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

type Matcher struct {
	fs             afero.Fs
	ownersFileName string
	ownersFiles    map[string]*OwnersFile
}

func NewMatcher(ownersFileName string) *Matcher {
	return newMatcherWithFs(ownersFileName, afero.NewOsFs())
}

func newMatcherWithFs(ownersFileName string, fs afero.Fs) *Matcher {
	return &Matcher{
		fs:             fs,
		ownersFileName: ownersFileName,
		ownersFiles:    make(map[string]*OwnersFile),
	}
}

func (m *Matcher) Load(dirPath string) (*OwnersFile, error) {
	dirPath = filepath.Clean(dirPath)
	if _, ok := m.ownersFiles[dirPath]; !ok {
		ownersFilePath := filepath.Join(dirPath, m.ownersFileName)
		if _, err := m.fs.Stat(ownersFilePath); err == nil {
			file, err := m.fs.Open(ownersFilePath)
			if err != nil {
				return nil, err
			}
			ownersFile, err := ParseFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse file %s: %w", ownersFilePath, err)
			}
			m.ownersFiles[dirPath] = ownersFile
		} else if errors.Is(err, os.ErrNotExist) {
			// Use an empty owners file struct if no file exists.
			m.ownersFiles[dirPath] = &OwnersFile{}
		} else if err != nil {
			return nil, err
		}
	}
	return m.ownersFiles[dirPath], nil
}

type MatchOwner struct {
	Owner    string
	Optional bool
}

func (m *Matcher) Match(filePath string) ([]MatchOwner, error) {
	// Search in a/b/OWNERS -> a/OWNERS -> OWNERS
	parts := strings.Split(filepath.Clean(filePath), string(os.PathSeparator))
	for i := len(parts) - 1; i >= 0; i-- {
		dirPath := filepath.Join(parts[:i]...)
		ownersFile, err := m.Load(dirPath)
		if err != nil {
			return nil, err
		}

		relFilePath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			return nil, err
		}

		matchedOwners, err := matchInFile(ownersFile, relFilePath)
		if err != nil {
			return nil, err
		}

		if len(matchedOwners) > 0 {
			return matchedOwners, nil
		}
	}
	return nil, nil
}

func matchInFile(ownersFile *OwnersFile, relFilePath string) ([]MatchOwner, error) {
	ownersToRequired := make(map[string]bool)
	for _, section := range ownersFile.Sections {
		for i := len(section.Rules) - 1; i >= 0; i-- {
			rule := section.Rules[i]

			matched, err := doublestar.PathMatch(rule.Pattern, relFilePath)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}

			owners := rule.Owners
			if len(owners) == 0 {
				owners = section.DefaultOwners
			}

			for _, owner := range owners {
				ownersToRequired[owner] = ownersToRequired[owner] || !section.Optional
			}

			break
		}
	}

	var sortedOwners []string
	for owner := range ownersToRequired {
		sortedOwners = append(sortedOwners, owner)
	}
	sort.Strings(sortedOwners)

	var matchedOwners []MatchOwner
	for _, owner := range sortedOwners {
		matchedOwners = append(matchedOwners, MatchOwner{
			Owner:    owner,
			Optional: !ownersToRequired[owner],
		})
	}

	return matchedOwners, nil
}
