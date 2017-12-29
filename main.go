package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	progPath string
	progName string

	rxBlame = regexp.MustCompile(`(?i:merge\s+(?:pull\s+request|pr)\s+#?(\d+)\s)`)
)

func init() {
	progPath = getProgramPath("git-blame-pr")
	progName = filepath.Base(progPath)

}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%s: No paths to blame\n", progName)
		os.Exit(1)
	}
	err := doBlame(os.Args[1:])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:error: %v\n", progName, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func doBlame(paths []string) error {
	var err error
	args := append([]string{"blame", "--first-parent"}, paths...)
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	srcPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	// Annotate PR id
	idHash := make(map[string]string)
	scanner := bufio.NewScanner(srcPipe)
	for scanner.Scan() {
		l := strings.SplitN(scanner.Text(), " ", 2)
		if len(l) == 2 {
			pr, ok := findPullRequst(idHash, l[0])
			if ok {
				fmt.Fprintf(os.Stdout, "PR #%-8s %s\n", pr, l[1])
				continue
			}
		}
		fmt.Println(scanner.Text()) // WHAT?
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func getProgramPath(def string) string {
	p, err := os.Executable()
	if err != nil {
		return def
	}
	return p
}

func findPullRequst(cache map[string]string, hash string) (string, bool) {
	result, ok := cache[hash]
	if ok {
		if len(result) == 0 {
			return result, false
		}
		return result, true
	}
	out, err := exec.Command("git", "show", "--oneline", hash).Output()
	if err != nil {
		cache[hash] = ""
		return hash, false
	}
	m := rxBlame.FindSubmatch(out)
	if len(m) < 2 {
		cache[hash] = ""
		return hash, false
	}
	pr := string(m[1])
	cache[hash] = pr
	return pr, true
}
