package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// use makes options/<letter>-*.md the live README.md.
func use(root, letter string) error {
	matches, _ := filepath.Glob(filepath.Join(root, "options", strings.ToLower(letter[:1])+"-*.md"))
	if len(matches) != 1 {
		all, _ := filepath.Glob(filepath.Join(root, "options", "?-*.md"))
		var names []string
		for _, m := range all {
			names = append(names, strings.TrimSuffix(filepath.Base(m), ".md"))
		}
		return fmt.Errorf("no option %q; pick one of %s", letter, strings.Join(names, ", "))
	}
	md, err := readText(matches[0])
	if err != nil {
		return err
	}
	readme := filepath.Join(root, "README.md")
	if err := writeText(root, readme, strings.ReplaceAll(md, "../assets/", "assets/")); err != nil {
		return err
	}
	fmt.Println("README.md <- options/" + filepath.Base(matches[0]))
	return nil
}
