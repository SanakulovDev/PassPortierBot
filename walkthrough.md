# Fix Undefined Variable `b` in main.go

The error `undefined: b` was caused by the bot startup logic being unreachable and out of scope inside the `retrievePassword` function.

## Changes

### cmd

#### [main.go](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/cmd/main.go)

- **Moved** `b.RemoveWebhook()` and `b.Start()` from the end of `retrievePassword` function back to the end of `main` function.

## Verification Results

### Automated Tests

- Ran `go build ./cmd/main.go` successfully (Exit Code 0).

# Optimize CI Workflow (`review.yml`)

Optimized the GitHub Actions workflow to be cleaner and more efficient.

## Changes

### .github/workflows

#### [review.yml](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/.github/workflows/review.yml)

- **Removed** duplicate `golangci-lint` step.
- **Replaced** `go mod tidy` with `go mod download` for safer CI execution.
- **Ensured** Go version is explicitly `1.24`.

# Fix Linter Errors (`errcheck`)

Fixed unchecked error return values identified by `errcheck` linter.

## Changes

### internal/storage

#### [db.go](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/internal/storage/db.go)

- Added `panic` if `db.AutoMigrate` fails, as this is a critical startup error.

### cmd

#### [main.go](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/cmd/main.go)

- Added error logging for `b.Delete` calls. Failed message deletion is treated as a warning.

# Multi-line Input & Security Auto-delete

Enhanced the input parsing to support multi-line credentials and added security measures to delete user messages.

## Changes

### cmd

#### [main.go](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/cmd/main.go)

- **Parsing**: Now supports `#service\npassword` format by splitting on the first whitespace sequence (space, newline, etc.) using `regexp`.
- **Auto-Delete**: Added logic to automatically delete the user's message (`defer b.Delete(c.Message())`) immediately triggering the handler, ensuring sensitive data is removed from chat history.

# Documentation Updates

Updated `README.md` to accurately reflect the current feature set.

## Changes

###

#### [README.md](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/README.md)

- **Removed** outdated references to AI/OCR/Voice features.
- **Added** documentation for:
  - **Zero-Knowledge** architecture.
  - **Manual Key Management System** (`#service password`).
  - **Multi-line Input** (`#service\npassword`).
  - **Auto-Delete** security feature.
  - Updated Usage Guide and Tech Stack.

#### [SECURITY.md](file:///Users/sanakulov/Documents/myProjects/PassPortierBot/SECURITY.md)

- **Created** security policy document containing:
  - Supported versions table.
  - Technical security details (Zero-Knowledge, Encryption types).
  - Vulnerability reporting contacts.
