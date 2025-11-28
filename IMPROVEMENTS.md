# Future Improvements & Extensions

This document outlines potential enhancements to take the shell to the next level.

## 🎯 Core Shell Features (Recommended)

These are standard features expected in any POSIX-compliant shell.

### 1. Globbing (Wildcard Expansion)

**What**: Support `*` and `?` patterns for file matching.

**Examples**:
```sh
$ ls *.go
$ echo app/*.go
$ cat test?.txt
```

**Implementation**:
- Use Go's `path/filepath.Glob` function during argument parsing
- Expand patterns before passing to commands
- Handle cases where no files match

**Files to modify**:
- `app/command.go`: Add glob expansion in `parseInput` or `parseQuotedArgs`

---

### 2. Environment Variable Expansion

**What**: Support `$VAR` syntax for variable substitution.

**Examples**:
```sh
$ echo $HOME
$ cd $MY_PROJECT
$ echo "User: $USER, Path: $PATH"
```

**Implementation**:
- Parse arguments for `$` prefix
- Use `os.Getenv` to replace with values
- Handle `${VAR}` syntax for clarity
- Support escaping with `\$`

**Files to modify**:
- `app/command.go`: Add variable expansion in `parseQuotedArgs`

---

### 3. Exit Code Tracking

**What**: Track the exit code of the last command in `$?`.

**Examples**:
```sh
$ ls /nonexistent
$ echo $?  # Should print non-zero
$ echo "success"
$ echo $?  # Should print 0
```

**Implementation**:
- Add `lastExitCode int` field to `Shell` struct
- Update after each command execution
- Expand `$?` during variable expansion

**Files to modify**:
- `app/shell.go`: Add field to `Shell` struct
- `app/command.go`: Update exit code after execution

---

## 🏗️ Advanced Features

### 4. Conditional Execution

**What**: Support `&&` (AND) and `||` (OR) operators.

**Examples**:
```sh
$ make && ./run_tests
$ command || echo "Failed!"
$ test -f file.txt && cat file.txt || echo "File not found"
```

**Implementation**:
- Extend `Command` struct with operator type
- Parse `&&` and `||` in `parseInput`
- Execute next command based on previous exit code

**Complexity**: Medium

---

### 5. Background Jobs

**What**: Support `&` for background processes and job control.

**Examples**:
```sh
$ long_running_task &
[1] 12345
$ jobs
[1]+  Running    long_running_task &
$ fg %1
$ bg %1
```

**Implementation**:
- Track running processes in `Shell` struct
- Implement `jobs`, `fg`, `bg` builtins
- Handle process groups and signals

**Complexity**: High

---

### 6. Subshells and Command Substitution

**What**: Support `$(command)` and backticks for command substitution.

**Examples**:
```sh
$ echo "Today is $(date)"
$ files=$(ls *.txt)
$ echo "Count: $(echo $files | wc -w)"
```

**Implementation**:
- Detect `$(...)` patterns
- Execute command in subshell
- Capture output and substitute

**Complexity**: Medium-High

---

## 🎨 User Experience Improvements

### 7. Signal Handling (Ctrl+C)

**What**: Properly handle interrupt signals without killing the shell.

**Current Issue**: Ctrl+C might kill the shell itself.

**Implementation**:
```go
import "os/signal"

// In NewShell or Run
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt)
go func() {
    for range sigChan {
        // Forward to running child process if any
        // Otherwise, just print new prompt
    }
}()
```

**Complexity**: Low

---

### 8. Customizable Prompt

**What**: Allow prompt customization via `PS1` environment variable.

**Examples**:
```sh
$ export PS1="[\u@\h \w]$ "
[user@hostname /current/dir]$ 
```

**Implementation**:
- Read `PS1` from environment
- Support escape sequences: `\u` (user), `\h` (hostname), `\w` (working dir)
- Update `readline.Config.Prompt` dynamically

**Complexity**: Low

---

### 9. Enhanced Autocomplete

**What**: Context-aware completion (files, directories, command arguments).

**Current**: Only completes command names.

**Improvements**:
- Complete file paths after `cd`, `cat`, etc.
- Complete options/flags for known commands
- Complete environment variables after `$`

**Complexity**: Medium

---

## 🧪 Testing & Quality

### 10. Integration Tests

**What**: End-to-end tests that run actual shell sessions.

**Example**:
```go
func TestShell_Integration_PipeAndRedirect(t *testing.T) {
    // Start shell, send commands, verify output
}
```

**Complexity**: Low

---

### 11. Benchmarks

**What**: Performance benchmarks for critical paths.

**Example**:
```go
func BenchmarkParseInput(b *testing.B) {
    shell := NewShell()
    for i := 0; i < b.N; i++ {
        shell.parseInput("echo hello | cat | wc")
    }
}
```

**Complexity**: Low

---

## 📊 Priority Recommendations

### Quick Wins (1-2 hours each)
1. ✅ Signal Handling (Ctrl+C)
2. ✅ Customizable Prompt (PS1)
3. ✅ Exit Code Tracking ($?)

### Medium Effort (3-5 hours each)
1. 🌟 **Globbing** - Most impactful for usability
2. 🌟 **Environment Variables** - Essential for scripting
3. Conditional Execution (&&, ||)

### Advanced Projects (1-2 days each)
1. Background Jobs & Job Control
2. Command Substitution
3. Enhanced Autocomplete

---

## 🚀 Getting Started

Pick one feature and:

1. **Create a branch**: `git checkout -b feature/globbing`
2. **Write tests first**: Add test cases in appropriate `*_test.go`
3. **Implement**: Make changes to source files
4. **Verify**: Run `go test ./app/...` and `codecrafters test`
5. **Document**: Update README with new feature
6. **Commit**: Use conventional commits (e.g., `feat: add globbing support`)

---

## 📚 Resources

- [POSIX Shell Specification](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html)
- [Bash Reference Manual](https://www.gnu.org/software/bash/manual/bash.html)
- [Go filepath.Glob](https://pkg.go.dev/path/filepath#Glob)
- [Go os/signal](https://pkg.go.dev/os/signal)
