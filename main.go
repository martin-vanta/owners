package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var verbose io.Writer = os.Stderr

func main() {
	if err := testableMain(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)

		if isRateLimitErr(err) {
			fmt.Println("Github API limit reached. Soft exiting")
		} else {
			os.Exit(1)
		}
	}
}

func testableMain(stdout io.Writer, args []string) error {
	opts, err := getOptions(stdout, args)
	if err != nil {
		return err
	}

	if opts == nil {
		return nil
	}

	commits := opts.baseRef + "..." + opts.headRef
	diff, err := run("git", "-C", opts.cwd, "diff", "--name-only", commits)
	if err != nil {
		return fmt.Errorf("error diffing %s: %w", commits, err)
	}

	paths, err := readLines(diff)
	if err != nil {
		return fmt.Errorf("error scanning lines from diff: %s\n%s", err, string(diff))
	}

	notifs, err := notifications(&gitfs{cwd: opts.cwd, rev: opts.baseRef}, paths, opts.filename)
	if err != nil {
		return err
	}

	if opts.author != "" {
		fmt.Fprintf(verbose, "not notifying pull request author %s\n", opts.author)
		delete(notifs, opts.author)
	}

	return opts.print(notifs)
}

func run(command string, args ...string) ([]byte, error) {
	out, err := exec.Command(command, args...).CombinedOutput()
	cmd := strings.Join(append([]string{command}, args...), " ")
	if err != nil {
		return nil, fmt.Errorf("error running command: %s -> %w", cmd, err)
	}
	fmt.Fprintln(verbose, cmd)
	fmt.Fprintln(verbose, string(out))
	return out, nil
}

func getOptions(stdout io.Writer, args []string) (*options, error) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return githubActionOptions()
	}
	return cliOptions(stdout, args)
}

func cliOptions(stdout io.Writer, args []string) (*options, error) {
	flags := flag.NewFlagSet("codenotify", flag.ContinueOnError)
	opts := options{}
	flags.StringVar(&opts.cwd, "cwd", "", "The working directory to use.")
	flags.StringVar(&opts.baseRef, "baseRef", "", "The base ref to use when computing the file diff.")
	flags.StringVar(&opts.headRef, "headRef", "HEAD", "The head ref to use when computing the file diff.")
	flags.StringVar(&opts.author, "author", "", "The author of the diff.")
	flags.StringVar(&opts.format, "format", "text", "The format of the output: text or markdown")
	flags.StringVar(&opts.filename, "filename", "CODENOTIFY", "The filename in which file subscribers are defined")
	flags.IntVar(&opts.subscriberThreshold, "subscriber-threshold", 0, "The threshold of notifying subscribers")
	v := *flags.Bool("verbose", false, "Verbose messages printed to stderr")

	if v {
		verbose = os.Stderr
	} else {
		verbose = io.Discard
	}

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	opts.print = func(notifs map[string][]string) error {
		return opts.writeNotifications(stdout, notifs)
	}
	return &opts, nil
}
