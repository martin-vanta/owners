package owners

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GenerateCodeOwners(ownersFileName, codeOwnersFilePath string) error {
	ownersFilePaths, err := findAllOwnersFiles(ownersFileName)
	if err != nil {
		return err
	}

	matcher := NewMatcher(ownersFileName)
	rules, err := getAllRequiredRules(matcher, ownersFilePaths)
	if err != nil {
		return err
	}

	var codeOwnersLines []string
	if _, err := os.Stat(codeOwnersFilePath); err == nil {
		codeOwnersLines, err = readLines(codeOwnersFilePath)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(codeOwnersFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeRequiredRules(f, codeOwnersLines, rules)
}

func findAllOwnersFiles(ownersFileName string) ([]string, error) {
	stdout, err := run("git", "ls-files", ownersFileName, fmt.Sprintf("**/%s", ownersFileName))
	if err != nil {
		return nil, err
	}

	lines := strings.Fields(stdout)
	sort.Strings(lines)
	return lines, nil
}

func getAllRequiredRules(matcher *Matcher, ownersFilePaths []string) ([]*Rule, error) {
	var allRequiredRules []*Rule
	for _, ownersFilePath := range ownersFilePaths {
		ownersFileDir, _ := filepath.Split(ownersFilePath)
		ownersFile, err := matcher.Load(ownersFileDir)
		if err != nil {
			return nil, err
		}

		for _, section := range ownersFile.Sections {
			if section.Optional {
				continue
			}
			for _, rule := range section.Rules {
				owners := rule.Owners
				if len(owners) == 0 {
					owners = section.DefaultOwners
				}
				allRequiredRules = append(allRequiredRules, &Rule{
					Pattern: filepath.Clean(filepath.Join(ownersFileDir, rule.Pattern)),
					Owners:  owners,
				})
			}
		}
	}
	return allRequiredRules, nil
}

const (
	headerLine = "# Generated by owners tool - do not edit below this line!"
	footerLine = "# Generated by owners tool - do not edit above this line!"

	generateStatusInit int = iota
	generateStatusFound
	generateStatusEnd
)

func writeRequiredRules(w io.Writer, codeOwnersLines []string, rules []*Rule) error {
	writer := bufio.NewWriter(w)

	var writeErr error
	writeLine := func(line string) {
		if writeErr != nil {
			return
		}
		_, err := writer.WriteString(line)
		if err != nil && writeErr == nil {
			writeErr = err
		}
		_, err = writer.WriteRune('\n')
		if err != nil && writeErr == nil {
			writeErr = err
		}
	}

	writeRules := func() {
		writeLine(headerLine)
		for _, rule := range rules {
			writeLine(fmt.Sprintf("%s %s", rule.Pattern, strings.Join(rule.Owners, " ")))
		}
		writeLine(footerLine)
	}

	status := generateStatusInit
	for _, line := range codeOwnersLines {
		if status == generateStatusFound {
			switch line {
			case "":
				status = generateStatusEnd
			case footerLine:
				status = generateStatusEnd
				continue
			default:
				continue
			}
		}

		if status == generateStatusInit && line == headerLine {
			writeRules()
			status = generateStatusFound
			continue
		}

		writeLine(line)
	}

	if status == generateStatusInit {
		writeLine("")
		writeRules()
	}

	if writeErr != nil {
		return writeErr
	}
	return writer.Flush()
}
