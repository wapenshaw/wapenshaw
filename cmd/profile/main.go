// Command profile builds the GitHub profile README for wapenshaw.
//
//	go run ./cmd/profile headers      regenerate the animated headers in assets/
//	go run ./cmd/profile refresh      refresh telemetry and upstream PR cards from public GitHub data
//	go run ./cmd/profile use <a|b|c>  make one of options/*.md the live README.md
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, err := repoRoot()
	if err == nil {
		err = run(root, os.Args[1:])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "profile:", err)
		os.Exit(1)
	}
}

func run(root string, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: profile headers | refresh | use <a|b|c>")
	}
	switch args[0] {
	case "headers":
		return writeHeaders(filepath.Join(root, "assets"))
	case "refresh":
		return refresh(root)
	case "use":
		if len(args) < 2 {
			return errors.New("usage: profile use <a|b|c>")
		}
		return use(root, args[1])
	}
	return fmt.Errorf("unknown command %q", args[0])
}

// repoRoot walks up from the working directory to the folder holding go.mod.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run inside the wapenshaw repo")
		}
		dir = parent
	}
}

// readText returns a file's contents with Windows line endings folded to \n.
func readText(path string) (string, error) {
	b, err := os.ReadFile(path)
	return strings.ReplaceAll(string(b), "\r\n", "\n"), err
}

// writeText writes only on change, so the workflow commits only when something moved.
func writeText(root, path, text string) error {
	if old, err := readText(path); err == nil && old == text {
		return nil
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return err
	}
	rel, _ := filepath.Rel(root, path)
	fmt.Println("updated", filepath.ToSlash(rel))
	return nil
}
