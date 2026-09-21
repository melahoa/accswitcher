package buildmode

// OverwatchVersion is this build variant's own version number, used for its
// update check against melahoa/accswitcher's GitHub releases. It is tracked
// separately from build/config.yml's info.version, which numbers the main
// app's releases against a different repository entirely - reusing it here
// would compare against the wrong release history. Bump this to match the
// tag of each Overwatch build release.
const OverwatchVersion = "0.4.0"
