package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/generator/nixgen"
)

// newGenerator returns the manifest generator a command obtains its link-farm through: the
// prebuilt one for apply --manifest, otherwise nix (→ ADR-0055 §1, §5). Choosing the generator
// and injecting it into the engine is all the cmd layer does with it.
func newGenerator() generator.Generator {
	return newGeneratorTo(os.Stderr)
}

// newGeneratorTo is newGenerator writing the generator's diagnostics to w instead of stderr
// (apply --all's stage 1 prefixes each config's lines; → ADR-0039).
//
// apply --all keeps ignoring --manifest and applies the entrypoint's configs through nix, as it
// did before generators; rejecting the combination is the selection mechanism's job (→ ADR-0056).
func newGeneratorTo(w io.Writer) generator.Generator {
	if flagManifest != "" && !flagApplyAll {
		return &generator.Prebuilt{}
	}
	return nixgen.New(w, flagDebug)
}

// generatorEnv is the environment variable naming the manifest generator (→ ADR-0056 §1).
const generatorEnv = "LAYAT_GENERATOR"

// selectableGenerators are the values --generator / LAYAT_GENERATOR accept. prebuilt is chosen only
// by --manifest and is never one of them (→ ADR-0056 §2).
var selectableGenerators = []string{generator.NameNix}

// resolveGenerator decides the manifest generator's name by the precedence --generator flag >
// LAYAT_GENERATOR env > default nix (→ ADR-0056 §1). An empty value is "not specified" and passes
// on to the next step; a step that decides the name leaves the steps below unread. An unknown
// value is an inputError naming where it came from.
func resolveGenerator(flag, env, projectDir, userConfigDir string) (string, error) {
	if flag != "" {
		return checkGeneratorName(flag, "--generator")
	}
	if env != "" {
		return checkGeneratorName(env, generatorEnv)
	}
	return generator.NameNix, nil
}

// checkGeneratorName returns name if it is selectable, otherwise an inputError naming source.
func checkGeneratorName(name, source string) (string, error) {
	if slices.Contains(selectableGenerators, name) {
		return name, nil
	}
	return "", &inputError{err: fmt.Errorf("layat: unknown generator %q in %s (available: %s)",
		name, source, strings.Join(selectableGenerators, ", "))}
}

// linePrefixWriter puts prefix at the head of every line written through it. Each Write reaches w
// in a single call, so lines from writers running in parallel do not interleave mid-line.
type linePrefixWriter struct {
	w      io.Writer
	prefix string
	mid    bool // the last byte written was not a newline
}

func (p *linePrefixWriter) Write(b []byte) (int, error) {
	out := make([]byte, 0, len(b)+len(p.prefix))
	for _, c := range b {
		if !p.mid {
			out = append(out, p.prefix...)
			p.mid = true
		}
		out = append(out, c)
		if c == '\n' {
			p.mid = false
		}
	}
	if _, err := p.w.Write(out); err != nil {
		return 0, err
	}
	return len(b), nil
}
