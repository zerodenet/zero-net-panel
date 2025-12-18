package cli

var (
	// Version is the human-friendly release tag, injected via ldflags.
	Version = "dev"
	// Commit is the git commit short SHA, injected via ldflags.
	Commit = "unknown"
	// BuildDate is an ISO-8601 UTC timestamp, injected via ldflags.
	BuildDate = ""
)
