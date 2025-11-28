# POSIX-Compliant Shell in Go

[![progress-banner](https://backend.codecrafters.io/progress/shell/ae298821-1348-4caf-83c1-63ed1d5b8728)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

A fully-featured POSIX-compliant shell built in Go, originally developed as part of the [CodeCrafters "Build Your Own Shell" Challenge](https://app.codecrafters.io/courses/shell/overview).

## Features

- ✅ **Command Execution**: Run external programs and builtins
- ✅ **Builtin Commands**: `cd`, `pwd`, `echo`, `type`, `exit`, `history`
- ✅ **Pipes**: Chain commands with `|` operator
- ✅ **I/O Redirection**: Support for `>`, `>>`, `2>`, `2>>`
- ✅ **Command History**: Persistent history with `HISTFILE` support
- ✅ **Quoting**: Handle single quotes, double quotes, and escape sequences
- ✅ **Tab Completion**: Autocomplete commands from PATH

## Project Structure

```
app/
├── main.go          # Entry point
├── shell.go         # Shell struct, REPL loop, autocomplete
├── command.go       # Command parsing & execution
├── builtins.go      # Builtin command handlers
├── utils.go         # Helper functions & constants
└── *_test.go        # Comprehensive test suite
```

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Running the Shell

```sh
./your_program.sh
```

### Running Tests

```sh
go test ./app/...
```

### Testing with CodeCrafters

```sh
codecrafters test    # Run all stages
codecrafters submit  # Submit your solution
```

## Usage Examples

```sh
# Basic commands
$ pwd
$ cd /tmp
$ echo "Hello, World!"

# Pipes
$ echo "test" | cat | wc

# I/O Redirection
$ echo "log entry" >> log.txt
$ cat nonexistent 2> error.log

# History
$ history
$ history 10           # Show last 10 entries
$ history -r ~/.history  # Read from file
$ history -w ~/.history  # Write to file

# With HISTFILE
$ HISTFILE=~/.shell_history ./your_program.sh
```

## Development

### Code Organization

The codebase is split into focused modules:
- **`shell.go`**: Core shell initialization, REPL, autocomplete
- **`command.go`**: Parsing (quotes, pipes, redirects) & execution
- **`builtins.go`**: All builtin command implementations
- **`utils.go`**: Shared utilities and constants

### Adding a New Builtin

1. Add to `builtinCommands` map in `builtins.go`
2. Implement `handle<Command>` function
3. Add case in `runCommand` switch statement
4. Write tests in `builtins_test.go`

## Acknowledgments

Built as part of the [CodeCrafters](https://codecrafters.io) "Build Your Own Shell" challenge.
