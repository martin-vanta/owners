package owners

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type Differ interface {
	Diff() ([]string, error)
}

type gitDiffer struct {
	since string
}

func NewGitDiffer(since string) Differ {
	return gitDiffer{
		since: since,
	}
}

func (d gitDiffer) Diff() ([]string, error) {
	// Find all files changed since ancestor commit of the reference and HEAD.
	cmd := exec.Command("git", "diff", fmt.Sprintf("%s...", d.since), "--name-only")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing '%s'\n%s", strings.Join(cmd.Args, " "), stderr.String())
	}

	lines := strings.Fields(stdout.String())
	sort.Strings(lines)
	return lines, nil
}

type fileDiffer struct {
	filePath string
}

func NewFileDiffer(filePath string) Differ {
	return fileDiffer{
		filePath: filePath,
	}
}

func (d fileDiffer) Diff() ([]string, error) {
	lines, err := readLines(d.filePath)
	if err != nil {
		return nil, err
	}
	sort.Strings(lines)
	return lines, nil
}

func readLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

type literalDiffer []string

func NewLiteralDiffer(filePaths []string) Differ {
	return literalDiffer(filePaths)
}

func (d literalDiffer) Diff() ([]string, error) {
	return d, nil
}
