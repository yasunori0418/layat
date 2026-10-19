package generation

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/yasunori0418/layat/internal/paths"
)

// DateLayout is the timestamp layout ListLinks reports for a generation link's mtime.
const DateLayout = "2006-01-02 15:04:05"

// ListLinks scans the profile's sibling generation links read-only and returns them in
// ascending order. The current generation is the one the profile link names, or failing that
// the one resolving to the same path as the profile.
func ListLinks(profileLink string) ([]Generation, error) {
	nums, err := generationLinks(profileLink)
	if err != nil {
		return nil, err
	}
	current := currentGeneration(profileLink, nums)
	gens := make([]Generation, 0, len(nums))
	for _, n := range nums {
		date := ""
		if fi, err := os.Lstat(paths.GenerationLink(profileLink, n)); err == nil {
			date = fi.ModTime().Format(DateLayout)
		}
		gens = append(gens, Generation{Number: n, Date: date, Current: n == current})
	}
	return gens, nil
}

// generationNumber parses the generation number out of a sibling name of the form
// <profile>-<N>-link; ok is false for any other name.
func generationNumber(profileLink, name string) (n int, ok bool) {
	prefix := filepath.Base(profileLink) + "-"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, "-link") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(name, prefix), "-link"))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// generationLinks returns the generation numbers of the <profile>-<N>-link siblings in
// ascending order. A missing profileDir holds no generations.
func generationLinks(profileLink string) ([]int, error) {
	entries, err := os.ReadDir(filepath.Dir(profileLink))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var nums []int
	for _, e := range entries {
		if n, ok := generationNumber(profileLink, e.Name()); ok {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums, nil
}

// currentGeneration returns the current generation number among nums, or -1 when the profile
// link is absent or points at none of them.
func currentGeneration(profileLink string, nums []int) int {
	dest, err := os.Readlink(profileLink)
	if err != nil {
		return -1
	}
	if n, ok := generationNumber(profileLink, filepath.Base(dest)); ok {
		return n
	}
	// The profile names something else (a link-farm directly): match by resolved path, taking
	// the newest generation that resolves to the same place.
	resolved, err := filepath.EvalSymlinks(profileLink)
	if err != nil {
		return -1
	}
	current := -1
	for _, n := range nums {
		if r, err := filepath.EvalSymlinks(paths.GenerationLink(profileLink, n)); err == nil && r == resolved {
			current = n
		}
	}
	return current
}
