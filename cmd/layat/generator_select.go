package main

import (
	"io"
	"os"

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
