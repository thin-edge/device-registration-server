
# Release
release:
    goreleaser release --rm-dist --auto-snapshot

# Build a snapshot of packages
release-snapshot *ARGS:
    goreleaser release --rm-dist --snapshot {{ARGS}}

# Build a snapshot for local development
build-snapshot *ARGS:
    goreleaser build --rm-dist --snapshot {{ARGS}}

run *ARGS:
    go run main.go {{ARGS}}
