package owners

import (
	"fmt"
	"sort"
	"strings"
)

func FindOwners(ownersFileName string, filePaths []string) (FindResults, error) {
	ownerToFiles := make(map[MatchOwner][]string)
	matcher := NewMatcher(ownersFileName)
	for _, filePath := range filePaths {
		matchedOwners, err := matcher.Match(filePath)
		if err != nil {
			return FindResults{}, err
		}

		for _, matchedOwner := range matchedOwners {
			ownerToFiles[matchedOwner] = append(ownerToFiles[matchedOwner], filePath)
		}
	}

	var results FindResults
	for matchedOwner, filePaths := range ownerToFiles {
		sort.Strings(filePaths)
		results.Owners = append(results.Owners, FindResult{
			Owner:     matchedOwner.Owner,
			Optional:  matchedOwner.Optional,
			FilePaths: filePaths,
		})
	}

	sort.Slice(results.Owners, func(i, j int) bool {
		if results.Owners[i].Owner != results.Owners[j].Owner {
			return results.Owners[i].Owner < results.Owners[j].Owner
		}
		return !results.Owners[i].Optional
	})

	return results, nil
}

type FindResults struct {
	Owners []FindResult `json:"owners"`
}

type FindResult struct {
	Owner     string   `json:"owner"`
	Optional  bool     `json:"optional"`
	FilePaths []string `json:"files"`
}

func (r FindResults) String() string {
	var s strings.Builder

	writeLinef := func(indentLevel int, format string, args ...interface{}) {
		for i := 0; i < indentLevel; i++ {
			s.WriteString("  ")
		}
		s.WriteString(fmt.Sprintf(format, args...))
		s.WriteRune('\n')
	}

	writeLinef(0, "results:")
	for _, result := range r.Owners {
		var optional string
		if result.Optional {
			optional = " (optional)"
		}

		writeLinef(1, "%s%s:", result.Owner, optional)
		for _, filePath := range result.FilePaths {
			writeLinef(2, filePath)
		}
	}

	return s.String()
}
