//go:build !overwatch

package buildmode

// IsOverwatchBuild reports whether this binary was built as the Overwatch
// variant (single Battle.net + Steam account list, no other platforms).
func IsOverwatchBuild() bool {
	return false
}
