package owners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

const (
	commentHeader = "<!-- github.com/martin-vanta/owners:github_actions_bot -->"
)

type gitHubEvent struct {
	PullRequest struct {
		Base struct {
			Sha string `json:"sha"`
		} `json:"base"`
		Head struct {
			Sha string `json:"sha"`
		} `json:"head"`
		NodeID string `json:"node_id"`
		User   struct {
			Login string `json:"login"`
		} `json:"User"`
		Draft bool `json:"draft"`
	} `json:"pull_request"`
}

type GitHubActions struct {
	PullRequestNodeID string
	Draft             bool
	BaseRef           string
	HeadRef           string
	Workspace         string
	OwnersFileName    string
	MaxNumOwners      int
	MaxNumFiles       int
}

func GetGitHubActions() (*GitHubActions, error) {
	path := os.Getenv("GITHUB_EVENT_PATH")
	if path == "" {
		return nil, fmt.Errorf("env var GITHUB_EVENT_PATH not set")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read GitHub event json %s: %s", path, err)
	}

	event := &gitHubEvent{}
	if err := json.Unmarshal(data, event); err != nil {
		return nil, fmt.Errorf("unable to decode GitHub event: %s\n%s", err, string(data))
	}

	workspace := os.Getenv("GITHUB_WORKSPACE")
	if path == "" {
		return nil, fmt.Errorf("env var GITHUB_WORKSPACE not set")
	}

	ownersFileName := os.Getenv("INPUT_OWNERS_FILE_NAME")
	if ownersFileName == "" {
		return nil, fmt.Errorf("env var INPUT_OWNERS_FILE_NAME not set")
	}

	maxNumOwners, _ := strconv.Atoi(os.Getenv("INPUT_MAX_NUM_OWNERS"))
	maxNumFiles, _ := strconv.Atoi(os.Getenv("INPUT_MAX_NUM_FILES"))

	return &GitHubActions{
		PullRequestNodeID: event.PullRequest.NodeID,
		Draft:             event.PullRequest.Draft,
		BaseRef:           event.PullRequest.Base.Sha,
		HeadRef:           event.PullRequest.Head.Sha,
		Workspace:         workspace,
		OwnersFileName:    ownersFileName,
		MaxNumOwners:      maxNumOwners,
		MaxNumFiles:       maxNumFiles,
	}, nil
}

func (g *GitHubActions) Prepare() error {
	if err := os.Chdir(g.Workspace); err != nil {
		return err
	}

	commitCount, err := getCommitCount(g.PullRequestNodeID)
	if err != nil {
		return err
	}

	_, err = run("git", "-c", "protocol.version=2", "fetch", "--deepen", strconv.Itoa(commitCount))
	if err != nil {
		return err
	}

	return nil
}

func (g *GitHubActions) WriteComment(results FindResults) error {
	comment := g.writeComment(results)

	commentId, err := findExistingCommentId(g.PullRequestNodeID, g.OwnersFileName)
	if err != nil {
		return err
	}

	if commentId == "" {
		// No comment exists and we don't need to notify any owners, so skip commenting.
		if len(results.Owners) == 0 {
			return nil
		}
		return addComment(g.PullRequestNodeID, comment)
	}

	return updateComment(commentId, comment)
}

func (g *GitHubActions) writeComment(results FindResults) string {
	w := &strings.Builder{}

	writeLinef := func(format string, args ...interface{}) {
		w.WriteString(fmt.Sprintf(format, args...))
		w.WriteRune('\n')
	}

	writeLinef(commentHeader)
	writeLinef("[Owners](https://github.com/martin-vanta/owners): Notifying file owners in %s files for diff %s...%s.\n\n", g.OwnersFileName, g.BaseRef, g.HeadRef)
	if len(results.Owners) == 0 {
		writeLinef("No notifications.")
	} else {
		writeLinef("| Owner | Required | File(s) |")
		writeLinef("|-|-|-|")
		for _, owner := range results.Owners {
			var required string
			if !owner.Optional {
				required = "✅"
			}

			files := owner.FilePaths
			if len(files) > g.MaxNumFiles {
				files := make([]string, g.MaxNumFiles+1)
				copy(files, owner.FilePaths[:g.MaxNumFiles])
				files[g.MaxNumFiles] = "..."
			}

			writeLinef("| %s | %s | %s |", owner.Owner, required, strings.Join(files, "<br>"))
		}
	}

	return w.String()
}

func updateComment(id, body string) error {
	// fmt.Fprintf(verbose, "updating existing comment: %s\n", id)
	return graphql(`
		mutation UpdateComment ($id: ID!, $body: String!) {
			updateIssueComment(input: {
				id: $id
				body: $body
			}) {
				clientMutationId
			}
		}`,
		map[string]interface{}{
			"id":   id,
			"body": body,
		},
		nil,
	)
}

func addComment(subjectId, body string) error {
	// fmt.Fprintf(verbose, "adding comment to pr %s\n", subjectId)
	return graphql(`
		mutation AddComment ($subjectId: ID!, $body: String!) {
			addComment(input: {
				subjectId: $subjectId
				body: $body
			}) {
				clientMutationId
			}
		}`,
		map[string]interface{}{
			"subjectId": subjectId,
			"body":      body,
		},
		nil,
	)
}

func getCommitCount(prNodeID string) (int, error) {
	data := struct {
		Node struct {
			Commits struct {
				TotalCount int `json:"totalCount"`
			} `json:"commits"`
		} `json:"node"`
	}{}
	err := graphql(`
		query CommitCount ($nodeId: ID!) {
			node(id: $nodeId) {
				... on PullRequest {
					commits {
						totalCount
					}
				}
			}
		}`,
		map[string]interface{}{
			"nodeId": prNodeID,
		},
		&data,
	)

	return data.Node.Commits.TotalCount, err
}

func findExistingCommentId(prNodeID string, filename string) (string, error) {
	data := struct {
		Node struct {
			Comments struct {
				Nodes []struct {
					Id     string `json:"id"`
					Author struct {
						Login string `json:"login"`
					} `json:"author"`
					Body string `json:"body"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	}{}
	err := graphql(`
		query GetPullRequestComments ($nodeId: ID!) {
			node(id: $nodeId) {
				... on PullRequest {
					comments(first: 100) {
						nodes {
							id
							author {
								login
							}
							body
						}
					}
				}
			}
		}`,
		map[string]interface{}{
			"nodeId": prNodeID,
		},
		&data,
	)
	if err != nil {
		return "", err
	}

	for _, comment := range data.Node.Comments.Nodes {
		if strings.HasPrefix(comment.Body, commentHeader) {
			return comment.Id, nil
		}
	}

	return "", nil
}

func graphql(query string, variables map[string]interface{}, responseData interface{}) error {
	reqbody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal query %s and variables %s: %w", query, variables, err)
	}

	req, err := http.NewRequest(http.MethodPost, os.Getenv("GITHUB_GRAPHQL_URL"), bytes.NewBuffer(reqbody))
	if err != nil {
		return err
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is not set")
	}
	req.Header.Set("Authorization", "bearer "+token)

	reqdump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Errorf("error dumping request: %w", err)
	}

	cl := &http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respdump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("error dumping response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response:\n%s\n\nrequest:\n%s", string(respdump), string(reqdump))
	}

	response := struct {
		Data   interface{}
		Errors []struct {
			Type    string   `json:"type"`
			Path    []string `json:"path"`
			Message string   `json:"message"`
		} `json:"errors"`
	}{
		Data: responseData,
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding json response:\n%s\n%w", respdump, err)
	}

	if len(response.Errors) > 0 {
		return fmt.Errorf("graphql error: %s\nrequest:\n%s", response.Errors[0].Message, reqdump)
	}

	return nil
}
