package owners

// import (
// 	"bytes"
// 	"errors"
// 	"os"
// 	"os/exec"
// )

// type FS interface {
// 	Open(name string) (File, error)
// }

// type File interface {
// 	Stat() (os.FileInfo, error)
// 	Read([]byte) (int, error)
// 	Close() error
// }

// // memfile is an in-memory file
// type memfile struct {
// 	*bytes.Buffer
// }

// func (m memfile) Close() error {
// 	return nil
// }

// func (m memfile) Stat() (os.FileInfo, error) {
// 	return nil, errors.New("memfile does not support stat")
// }

// // gitfs implements the FS interface for files at a specific git revision.
// type gitfs struct {
// 	cwd string
// 	rev string
// }

// func (g *gitfs) Open(name string) (File, error) {
// 	cmd := exec.Command("git", "-C", g.cwd, "show", g.rev+":"+name)
// 	buf, err := cmd.Output()
// 	if err != nil {
// 		return nil, os.ErrNotExist
// 	}
// 	return memfile{
// 		Buffer: bytes.NewBuffer(buf),
// 	}, nil
// }
