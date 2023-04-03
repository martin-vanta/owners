// Links:
// Docs: https://docs.gitlab.com/ee/user/project/code_owners.html
// Reference Impl: https://gitlab.com/gitlab-org/gitlab/-/tree/master/ee/lib/gitlab/code_owners

package owners

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultSectionName = "OWNERS"
)

type RuleFile struct {
	Path     string
	Sections []Section
}

type Section struct {
	Name          string
	Optional      bool
	Approvals     int
	DefaultOwners []string
	Rules         []Rule
}

type Rule struct {
	Pattern string
}

func ParseFile(r io.Reader) (*RuleFile, error) {
	file := &RuleFile{}
	// currSection := Section{Name: defaultSectionName}

	var lineno int
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		lineno++

		// Strip comments
		commentSplit := strings.SplitN(line, "#", 2)
		line = strings.TrimSpace(commentSplit[0])

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return file, nil
}

var (
	// e.g. ^[Documentation][2] @docs-team
	sectionHeaderRe = regexp.MustCompile(strings.Join([]string{
		`^`,
		`(?P<optional>\^)?`,
		`\[(?P<name>\w+)\]`,
		`(\[(?P<approvals>-?\d+)\])?`,
		`(?P<default_owners>(\s@\w+(/\w+)?)*)`,
		`$`,
	}, ""))
)

func parseSectionHeader(line string) *Section {
	matches := sectionHeaderRe.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nil
	}

	section := &Section{Approvals: 1}
	for i, groupName := range sectionHeaderRe.SubexpNames() {
		match := matches[i]
		if match == "" {
			continue
		}

		switch groupName {
		case "name":
			section.Name = match
		case "optional":
			section.Optional = match == "^"
		case "approvals":
			approvals, err := strconv.ParseInt(match, 10, 64)
			if err != nil || approvals < 1 {
				approvals = 1
			}
			section.Approvals = int(approvals)
		case "default_owners":
			section.DefaultOwners = strings.Fields(match)
		}
	}
	return section
}
