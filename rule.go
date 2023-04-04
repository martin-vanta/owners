package owners

// import (
// 	"bufio"
// 	"bytes"
// 	"fmt"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"strings"
// )

// func subscribers(fs FS, path string, notifyFilename string) ([]string, error) {
// 	fmt.Fprintf(verbose, "analyzing subscribers in %s files\n", notifyFilename)
// 	subscribers := []string{}

// 	parts := strings.Split(path, string(os.PathSeparator))
// 	for i := range parts {
// 		base := filepath.Join(parts[:i]...)
// 		rulefilepath := filepath.Join(base, notifyFilename)

// 		rulefile, err := fs.Open(rulefilepath)
// 		if err != nil {
// 			if err == os.ErrNotExist {
// 				continue
// 			}
// 			return nil, err
// 		}

// 		scanner := bufio.NewScanner(rulefile)
// 		for scanner.Scan() {
// 			rule := scanner.Text()
// 			if rule != "" && rule[0] == '#' {
// 				// skip comment
// 				continue
// 			}

// 			fields := strings.Fields(rule)
// 			switch len(fields) {
// 			case 0:
// 				// skip blank line
// 				continue
// 			case 1:
// 				return nil, fmt.Errorf("expected at least two fields for rule in %s: %s", rulefilepath, rule)
// 			}

// 			rel, err := filepath.Rel(base, path)
// 			if err != nil {
// 				return nil, err
// 			}

// 			re, err := patternToRegexp(fields[0])
// 			if err != nil {
// 				return nil, fmt.Errorf("invalid pattern in %s: %s: %w", rulefilepath, rule, err)
// 			}

// 			if re.MatchString(rel) {
// 				subscribers = append(subscribers, fields[1:]...)
// 			}
// 		}

// 		if err := scanner.Err(); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return subscribers, nil
// }

// func patternToRegexp(pattern string) (*regexp.Regexp, error) {
// 	if pattern[len(pattern)-1:] == "/" {
// 		pattern += "**"
// 	}
// 	pattern = regexp.QuoteMeta(pattern)
// 	pattern = strings.ReplaceAll(pattern, `/\*\*/`, "/([^/]*/)*")
// 	pattern = strings.ReplaceAll(pattern, `\*\*/`, "([^/]+/)*")
// 	pattern = strings.ReplaceAll(pattern, `/\*\*`, ".*")
// 	pattern = strings.ReplaceAll(pattern, `\*\*`, ".*")
// 	pattern = strings.ReplaceAll(pattern, `\*`, "[^/]*")
// 	pattern = "^" + pattern + "$"
// 	return regexp.Compile(pattern)
// }

// func readLinesBytes(b []byte) ([]string, error) {
// 	lines := []string{}
// 	scanner := bufio.NewScanner(bytes.NewBuffer(b))
// 	for scanner.Scan() {
// 		lines = append(lines, scanner.Text())
// 	}
// 	return lines, scanner.Err()
// }
