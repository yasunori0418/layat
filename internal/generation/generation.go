// Package generation is the contract between the engine and the generation backend: the
// operations that record, list, and switch generations of a profile, and the bases the
// profileDirs live under. Inputs and outputs are paths and integers only.
package generation

// Generation is one generation of a profile.
type Generation struct {
	Number  int    // generation number
	Date    string // creation timestamp (the backend's display verbatim; carried unparsed)
	Current bool   // whether it is the current generation (the one the profile link points at)
}

// Backend records, lists, and switches the generations of a profile and owns the bases the
// profileDirs are keyed under. profileLink is <profileDir>/profile; the backend keeps the
// sibling layout <profileDir>/profile-<N>-link the engine reads.
type Backend interface {
	// Commit records linkFarm as a new generation of profileLink and points the profile at it.
	Commit(profileLink, linkFarm string) error
	// List returns the generations of profileLink in ascending order, marking the current one.
	List(profileLink string) ([]Generation, error)
	// Switch points profileLink at the existing generation gen.
	Switch(profileLink string, gen int) error
	// UserBase returns the base the user-state profileDirs are keyed under for <state>.
	UserBase(stateDir string) string
	// SystemBase returns the base the system-mode profileDirs are keyed under.
	SystemBase() string
}
