package owners

import (
	"errors"
	"os"
	"path"

	"github.com/spf13/afero"
)

const (
	defaultOwnersFileName = "OWNERS"
)

type Loader struct {
	fs             afero.Fs // Mocked FS for testing.
	ownersFileName string
	ownersFiles    map[string]*OwnersFile
}

func NewLoader(ownersFileName string) *Loader {
	return newLoaderWithFs(ownersFileName, afero.NewOsFs())
}

func newLoaderWithFs(ownersFileName string, fs afero.Fs) *Loader {
	return &Loader{
		fs:             fs,
		ownersFileName: ownersFileName,
		ownersFiles:    make(map[string]*OwnersFile),
	}
}

func (l *Loader) Load(dirPath string) (*OwnersFile, error) {
	if _, ok := l.ownersFiles[dirPath]; !ok {
		ownersFilePath := path.Join(dirPath, l.ownersFileName)
		if _, err := l.fs.Stat(ownersFilePath); err == nil {
			file, err := l.fs.Open(ownersFilePath)
			if err != nil {
				return nil, err
			}
			ownersFile, err := ParseFile(file)
			if err != nil {
				return nil, err
			}
			l.ownersFiles[dirPath] = ownersFile
		} else if errors.Is(err, os.ErrNotExist) {
			l.ownersFiles[dirPath] = &OwnersFile{}
		} else if err != nil {
			return nil, err
		}
	}
	return l.ownersFiles[dirPath], nil
}
