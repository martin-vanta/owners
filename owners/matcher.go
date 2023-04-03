package owners

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/afero"
)

type Matcher struct {
	fs             afero.Fs // Mocked FS for testing.
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
				return nil, err
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

func (m *Matcher) Match(filePath string) (*Match, error) {
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

		match, err := matchInFile(ownersFile, relFilePath)
		if err != nil {
			return nil, err
		}

		if match != nil {
			return match, nil
		}
	}

	return nil, nil
}

func matchInFile(ownersFile *OwnersFile, relFilePath string) (*Match, error) {
	match := &Match{}
	for _, section := range ownersFile.Sections {
		for i := len(section.Rules) - 1; i >= 0; i-- {
			rule := section.Rules[i]

			matched, err := doublestar.PathMatch(rule.Pattern, relFilePath)
			if err != nil {
				return nil, err
			}

			if matched {
				owners := rule.Owners
				if len(owners) == 0 {
					owners = section.DefaultOwners
				}
				if section.Optional {
					match.OptionalOwners = append(match.OptionalOwners, owners...)
				} else {
					match.RequiredOwners = append(match.RequiredOwners, owners...)
				}
				break
			}
		}
	}

	if len(match.RequiredOwners) == 0 && len(match.OptionalOwners) == 0 {
		return nil, nil
	}
	return match, nil
}
