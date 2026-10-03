package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/yasunori0418/layat/internal/generator"
	"github.com/yasunori0418/layat/internal/generator/nixgen"
)

// newGenerator returns the manifest generator a command obtains its link-farm through, chosen by
// selectGenerator: the prebuilt one for apply --manifest, otherwise the one the flag, environment
// or settings name (→ ADR-0055 §1, §5, ADR-0056). Choosing the generator and injecting it into the
// engine is all the cmd layer does with it. A rejected selection is an inputError.
func newGenerator() (generator.Generator, error) {
	name, err := selectGenerator()
	if err != nil {
		return nil, err
	}
	return newGeneratorTo(name, os.Stderr), nil
}

// newGeneratorTo returns the generator named name (already selected) writing its diagnostics to
// w (apply --all's stage 1 resolves the name once and prefixes each config's lines; → ADR-0039).
func newGeneratorTo(name string, w io.Writer) generator.Generator {
	if name == generator.NamePrebuilt {
		return &generator.Prebuilt{}
	}
	return nixgen.New(w, flagDebug)
}

// selectGenerator chooses the generator's name for the running command. apply --manifest takes
// prebuilt without reading the environment or any settings file (module activation runs in an
// environment the user does not control) and rejects --generator; otherwise resolveGenerator
// decides from the flag, LAYAT_GENERATOR, the project setting and the user setting (→ ADR-0056).
func selectGenerator() (string, error) {
	if flagManifest != "" {
		if flagGenerator != "" {
			return "", &inputError{err: errors.New("layat: --manifest cannot be combined with --generator (--manifest fixes the source to a pre-built link-farm)")}
		}
		return generator.NamePrebuilt, nil
	}
	projectDir, userConfigDir := projectSettingDir(flagFile), userSettingDir()
	if flagFile != "" && projectDir == "" {
		// An -f path that cannot be stat'ed reads no settings file; the generator's discovery
		// reports the missing entrypoint (→ ADR-0056 §3).
		userConfigDir = ""
	}
	return resolveGenerator(flagGenerator, os.Getenv(generatorEnv), projectDir, userConfigDir)
}

// projectSettingDir is the directory holding the project setting layat.toml: the -f directory
// (the file's directory when -f names a file), otherwise the CWD. Parent directories are not
// searched. It is "" when -f cannot be stat'ed or the CWD cannot be resolved.
func projectSettingDir(file string) string {
	if file == "" {
		wd, err := os.Getwd()
		if err != nil {
			return ""
		}
		return wd
	}
	fi, err := os.Stat(file)
	if err != nil {
		return ""
	}
	if fi.IsDir() {
		return file
	}
	return filepath.Dir(file)
}

// userSettingDir is the base directory of the user setting <dir>/layat/config.toml:
// $XDG_CONFIG_HOME, or ~/.config when it is unset. It is "" when neither can be resolved.
func userSettingDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

// generatorEnv is the environment variable naming the manifest generator (→ ADR-0056 §1).
const generatorEnv = "LAYAT_GENERATOR"

// selectableGenerators are the values --generator / LAYAT_GENERATOR accept. prebuilt is chosen only
// by --manifest and is never one of them (→ ADR-0056 §2).
var selectableGenerators = []string{generator.NameNix}

// generatorSetting is the schema of layat.toml / config.toml: generator alone, no version item.
// An unknown key is rejected (strict; → ADR-0056 §4).
type generatorSetting struct {
	Generator string `toml:"generator"`
}

// resolveGenerator decides the manifest generator's name by the precedence --generator flag >
// LAYAT_GENERATOR env > project setting <projectDir>/layat.toml > user setting
// <userConfigDir>/layat/config.toml > default nix (→ ADR-0056 §1). An empty value, a missing
// file, an empty directory and a file without the generator key are "not specified" and pass on to
// the next step; a step that decides the name leaves the steps below unread. An unknown value and
// a malformed settings file are an inputError naming where they came from.
func resolveGenerator(flag, env, projectDir, userConfigDir string) (string, error) {
	if flag != "" {
		return checkGeneratorName(flag, "--generator")
	}
	if env != "" {
		return checkGeneratorName(env, generatorEnv)
	}
	var files []string
	if projectDir != "" {
		files = append(files, filepath.Join(projectDir, "layat.toml"))
	}
	if userConfigDir != "" {
		files = append(files, filepath.Join(userConfigDir, "layat", "config.toml"))
	}
	for _, path := range files {
		name, err := readGeneratorSetting(path)
		if err != nil {
			return "", err
		}
		if name != "" {
			return checkGeneratorName(name, path)
		}
	}
	return generator.NameNix, nil
}

// readGeneratorSetting returns the generator a settings file names, "" when the file does not
// exist or has no generator key. A parse failure or an unknown key is an inputError; any other
// read failure is returned as is, keeping its own classification.
func readGeneratorSetting(path string) (string, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("layat: cannot read %s: %w", path, err)
	}
	var s generatorSetting
	err = toml.NewDecoder(bytes.NewReader(b)).DisallowUnknownFields().Decode(&s)
	var sme *toml.StrictMissingError
	if errors.As(err, &sme) {
		keys := make([]string, 0, len(sme.Errors))
		for _, e := range sme.Errors {
			keys = append(keys, fmt.Sprintf("%q", strings.Join(e.Key(), ".")))
		}
		return "", &inputError{err: fmt.Errorf("layat: %s: unknown key %s (only \"generator\" is allowed)",
			path, strings.Join(keys, ", "))}
	}
	if err != nil {
		return "", &inputError{err: fmt.Errorf("layat: %s: %w", path, err)}
	}
	return s.Generator, nil
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
