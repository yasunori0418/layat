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

// newGenerator returns the manifest generator chosen by selectGenerator. A rejected selection is
// an inputError.
func newGenerator() (generator.Generator, error) {
	name, err := selectGenerator()
	if err != nil {
		return nil, err
	}
	return newGeneratorTo(name, os.Stderr), nil
}

// newGeneratorTo returns the generator named name, writing its diagnostics to w.
func newGeneratorTo(name string, w io.Writer) generator.Generator {
	if name == generator.NamePrebuilt {
		if flagApplyAll {
			return &generator.Prebuilt{LinkFarms: manifestSet(flagManifest)}
		}
		return &generator.Prebuilt{}
	}
	return nixgen.New(w, flagDebug)
}

// selectGenerator chooses the generator's name. apply --manifest takes prebuilt without reading
// the environment or settings, rejecting --generator and a malformed --manifest; otherwise
// resolveGenerator decides.
func selectGenerator() (string, error) {
	if len(flagManifest) > 0 {
		if flagGenerator != "" {
			return "", &inputError{err: errors.New("layat: --manifest cannot be combined with --generator (--manifest fixes the source to a pre-built link-farm)")}
		}
		if err := checkManifestForm(flagManifest, flagApplyAll); err != nil {
			return "", err
		}
		return generator.NamePrebuilt, nil
	}
	projectDir, userConfigDir := projectSettingDir(flagFile), userSettingDir()
	if flagFile != "" && projectDir == "" {
		// An -f path that cannot be stat'ed reads no settings file; discovery reports it.
		userConfigDir = ""
	}
	return resolveGenerator(flagGenerator, os.Getenv(generatorEnv), projectDir, userConfigDir)
}

// checkManifestForm rejects a --manifest list that does not fit its mode: one bare link-farm path
// without --all, or <name>=<link-farm> pairs with non-empty, distinct names under --all.
func checkManifestForm(values []string, all bool) error {
	if !all {
		if len(values) > 1 {
			return &inputError{err: errors.New("layat: --manifest can be given only once without --all (repeat --manifest <name>=<link-farm> with --all)")}
		}
		if strings.Contains(values[0], "=") {
			return &inputError{err: fmt.Errorf("layat: --manifest %q: the <name>=<link-farm> form requires --all", values[0])}
		}
		return nil
	}
	seen := map[string]bool{}
	for _, v := range values {
		name, _, ok := strings.Cut(v, "=")
		switch {
		case !ok:
			return &inputError{err: fmt.Errorf("layat: --manifest %q: with --all, give each config as <name>=<link-farm>", v)}
		case name == "":
			return &inputError{err: fmt.Errorf("layat: --manifest %q: the config name before \"=\" is empty", v)}
		case seen[name]:
			return &inputError{err: fmt.Errorf("layat: --manifest: config %q is given more than once", name)}
		}
		seen[name] = true
	}
	return nil
}

// manifestSet maps each --manifest <name>=<link-farm> pair, split at the first "=", to its config.
func manifestSet(values []string) map[string]string {
	farms := make(map[string]string, len(values))
	for _, v := range values {
		name, farm, _ := strings.Cut(v, "=")
		farms[name] = farm
	}
	return farms
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

// generatorEnv is the environment variable naming the manifest generator.
const generatorEnv = "LAYAT_GENERATOR"

// selectableGenerators are the values --generator / LAYAT_GENERATOR accept. prebuilt is chosen
// only by --manifest.
var selectableGenerators = []string{generator.NameNix}

// generatorSetting is the schema of layat.toml / config.toml: generator alone. Unknown keys are
// rejected.
type generatorSetting struct {
	Generator string `toml:"generator"`
}

// resolveGenerator decides the generator by precedence: flag > env > <projectDir>/layat.toml >
// <userConfigDir>/layat/config.toml > nix. Unset values and missing files pass to the next step;
// an unknown value or malformed file is an inputError naming its source.
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
