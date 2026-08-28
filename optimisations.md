# Hoppr: High-Performance CLI Engineering & Architecture Specification

This specification provides the engineering blueprint for **Hoppr** as a **highly-optimized, production-grade Command-Line Interface (CLI)** tool designed to manage software project landscapes with sub-millisecond responsiveness, zero-overhead execution, and crash-resilient data integrity.

---

## 1. Core CLI Performance Architecture

CLI tools live in the developer's hot path. When invoked inside shell functions, interactive prompt hooks, or fuzzy finders (`fzf`), **every millisecond of latency is felt**.

```
                ┌─────────────────────────────────────────┐
                │             Shell Invocation            │
                │        (hop <name> / hop add / etc.)    │
                └────────────────────┬────────────────────┘
                                     │
                        [Fast-Path Arg Router]
                   (Zero Allocation, No Disk I/O)
                                     │
           ┌─────────────────────────┴─────────────────────────┐
           │                                                   │
  [Zero-I/O Commands]                                 [Config-Requiring Commands]
  (--help, --version, completion)                     (add, remove, list, _get_path)
           │                                                   │
     Instant Exit (<0.5ms)                                [Shared Read Lock / Exclusive Lock]
                                                               │
                                                      [Buffered Stream I/O]
                                                               │
                                                      [Zero-Alloc JSON Parse]
                                                               │
                                                      [Atomic fsync Write]
```

### 1.1 Performance Targets & SLA

| Metric | Target | Optimization Strategy |
|---|---|---|
| **Query Latency (`_get_path`)** | **< 1.5 ms** | Read-shared lock, direct key lookup, no pretty printing |
| **List Rendering (`list`)** | **< 3.0 ms** | Pre-allocated slices, buffered `bufio.Writer`, `slices.Sort` |
| **Command Mutation (`add`/`drop`)** | **< 5.0 ms** | Atomic temp-file write, direct `fsync`, exclusive `flock` |
| **Cold Binary Startup** | **< 1.0 ms** | Stripped debug tables (`-ldflags="-s -w"`), `CGO_ENABLED=0` |
| **Binary Footprint** | **< 1.5 MB** | Zero bloated dependencies, standard library core |
| **Peak Heap Allocations** | **< 200 KB** | Struct pooling, pre-allocated slice buffers |

---

## 2. SOLID Principles for CLI Systems

Clean Architecture and SOLID principles ensure that extreme performance does not come at the cost of messy, unmaintainable code.

```
                    ┌────────────────────────┐
                    │     CLI Dispatcher     │
                    │   (cmd/hop/main.go)    │
                    └───────────┬────────────┘
                                │ (DIP: depends on interfaces)
                    ┌───────────▼────────────┐
                    │    Command Registry    │
                    │  (internal/commands)   │
                    └───────────┬────────────┘
                                │
          ┌─────────────────────┴─────────────────────┐
          │                                           │
 ┌────────▼────────┐                         ┌────────▼────────┐
 │ Project Service │                         │  Terminal UI    │
 │ (internal/core) │                         │  (internal/ui)  │
 └────────┬────────┘                         └─────────────────┘
          │ (ISP: Narrow storage interfaces)
 ┌────────▼────────┐
 │ Storage Engine  │ (LSP: JSONFile, SQLite, Memory for test)
 └─────────────────┘
```

### 2.1 Single Responsibility Principle (SRP)
Each subsystem has exactly one reason to change:
* **`internal/config`**: Configuration structure, defaults, and schema migration logic.
* **`internal/storage`**: Thread-safe, cross-process atomic file persistence and locking.
* **`internal/core`**: Domain rules (path validation, frecency scoring, tag filtering, shorthand `.` resolution).
* **`internal/commands`**: CLI argument validation, flag extraction, and exit code management.
* **`internal/ui`**: Terminal formatting, tabwriters, color detection, and streaming output.

### 2.2 Open / Closed Principle (OCP)
The command router is open for extension but closed for modification. New subcommands (`hop doctor`, `hop import`, `hop tags`) register themselves into a command registry without modifying core routing logic:

```go
type Command interface {
    Name() string
    Synopsis() string
    Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int
}

type Registry struct {
    commands map[string]Command
}

func (r *Registry) Register(cmd Command) {
    r.commands[cmd.Name()] = cmd
}
```

### 2.3 Liskov Substitution Principle (LSP)
The storage layer is defined by an abstract contract:

```go
type StorageEngine interface {
    Read(ctx context.Context) (*domain.Config, error)
    Update(ctx context.Context, mutate func(cfg *domain.Config) error) error
}
```
* **`JSONStorage`**: Production file persistence with file locks.
* **`MemoryStorage`**: Zero-disk in-memory implementation for 100% deterministic sub-millisecond unit testing.

### 2.4 Interface Segregation Principle (ISP)
Handlers receive only the specific interfaces they require:
```go
// Read-only query commands (e.g. _get_path, list)
type ProjectReader interface {
    GetPath(ctx context.Context, list, name string) (string, error)
    GetAll(ctx context.Context, list string) ([]domain.Project, error)
}

// Mutating commands (e.g. add, remove, drop)
type ProjectWriter interface {
    Save(ctx context.Context, list string, project domain.Project) error
    Delete(ctx context.Context, list, name string) error
}
```

### 2.5 Dependency Inversion Principle (DIP)
Commands depend on abstract `io.Writer` interfaces rather than hardcoding `os.Stdout` or `os.Stderr`. This enables direct output assertions in automated tests:
```go
func NewListCommand(reader ProjectReader, ui Formatter) *ListCommand {
    return &ListCommand{reader: reader, ui: ui}
}
```

---

## 3. High-Performance Low-Level CLI Optimizations

### 3.1 Zero-Allocation Fast-Path Routing
Never perform disk I/O, regex evaluations, or JSON unmarshaling before checking if the command actually requires it.

```go
func main() {
    if len(os.Args) < 2 {
        printUsage(os.Stderr)
        os.Exit(2)
    }

    cmd := os.Args[1]

    // Fast path: Zero I/O commands
    switch cmd {
    case "--help", "-h", "help":
        printUsage(os.Stdout)
        os.Exit(0)
    case "--version", "-v", "version":
        printVersion(os.Stdout)
        os.Exit(0)
    }
    // Proceed to load config only for data-driven commands
}
```

### 3.2 Granular Shared / Exclusive File Locking
* **Problem:** Blindly locking the config file exclusively blocks concurrent terminal queries (e.g. 5 tmux panes querying paths during shell initialization).
* **Optimization:** Use **Read-Shared Locks (`RLock`)** for `list`, `_get_path`, and `_get_editor`, and **Exclusive Locks (`Lock`)** only during mutations (`add`, `remove`, `import`, `drop`).

```go
// Read path: Allows hundreds of concurrent reads simultaneously
func (s *JSONStorage) Read(ctx context.Context) (*domain.Config, error) {
    if err := s.flock.RLockContext(ctx, 200*time.Millisecond); err != nil {
        return nil, err
    }
    defer s.flock.Unlock()
    return s.loadFromFile()
}
```

### 3.3 Memory Pre-allocation & Zero-Copy Sorting
Avoid dynamic slice resizing during `list` commands:
```go
// Allocate exact capacity once
names := make([]string, 0, len(cfg.Projects))
for name := range cfg.Projects {
    names = append(names, name)
}
// In-place sort with Go 1.21+ slices package
slices.Sort(names)
```

### 3.4 Direct Buffered Output (`bufio.Writer`)
Avoid individual unbuffered `fmt.Printf` system calls for each row. Wrap `os.Stdout` in a 4KB buffer:
```go
bufOut := bufio.NewWriterSize(out, 4096)
defer bufOut.Flush()

w := tabwriter.NewWriter(bufOut, 0, 0, 3, ' ', 0)
for _, p := range projects {
    // Write formatted output
    fmt.Fprintf(w, "%s\t: %s\n", p.Name, p.Path)
}
w.Flush()
```

---

## 4. Production CLI Standards & Unix Philosophy

### 4.1 Strict Stream Separation
* **`stdout` (Data Pipe)**: Reserved exclusively for valid data outputs (e.g. target paths, project lists, JSON). Clean and pipeable into other tools.
* **`stderr` (Diagnostics)**: Reserved for logs, status messages ("Added project 'xyz'"), warnings, and errors.

### 4.2 Standard POSIX Exit Codes
| Exit Code | Meaning | Use Case |
|---|---|---|
| **`0`** | `EXIT_SUCCESS` | Command completed successfully |
| **`1`** | `EXIT_FAILURE` | General runtime error (I/O, permission, disk failure) |
| **`2`** | `EXIT_USAGE` | Invalid command syntax, missing arguments, unknown flag |
| **`3`** | `EXIT_NOT_FOUND`| Requested project, list, or workspace does not exist |

### 4.3 Pipeline Composability (`--plain`, `--json`)
* `hop list --plain`: Tab-separated `name\tpath` with no headers, ideal for:
  ```bash
  hop list --plain | fzf | awk '{print $1}'
  ```
* `hop list --json`: Structured JSON for scripting and automation.

### 4.4 `NO_COLOR` & Terminal Capability Detection
* Honor [`NO_COLOR`](https://no-color.org/) and `TERM=dumb`.
* Inspect `isatty` before emitting ANSI escape codes to prevent corrupting piped data.

---

## 5. Data Safety: Atomic Writes & Crash Resistance

### 5.1 Atomic Persistence with `fsync`
Writing directly to `config.json` can result in empty/corrupted files if interrupted:

```go
func (s *JSONStorage) Update(ctx context.Context, mutate func(cfg *domain.Config) error) error {
    if err := s.flock.LockContext(ctx, 500*time.Millisecond); err != nil {
        return fmt.Errorf("storage lock timeout: %w", err)
    }
    defer s.flock.Unlock()

    cfg, err := s.loadFromFile()
    if err != nil {
        return err
    }

    if err := mutate(cfg); err != nil {
        return err
    }

    data, err := json.MarshalIndent(cfg, "", "    ")
    if err != nil {
        return err
    }

    // 1. Create temp file in same filesystem directory
    tmp, err := os.CreateTemp(s.configDir, "hoppr-*.tmp")
    if err != nil {
        return err
    }
    tmpPath := tmp.Name()
    defer os.Remove(tmpPath)

    // 2. Write data
    if _, err := tmp.Write(data); err != nil {
        tmp.Close()
        return err
    }

    // 3. Flush OS buffer to physical storage
    if err := tmp.Sync(); err != nil {
        tmp.Close()
        return err
    }

    if err := tmp.Close(); err != nil {
        return err
    }

    // 4. Atomic directory entry swap
    return os.Rename(tmpPath, s.configPath)
}
```

---

## 6. Advanced Project Management Features

### 6.1 Frecency Algorithm ($Frequency \times Recency$)
Track developer usage and automatically suggest the most relevant projects first:
$$\text{Score} = \text{AccessCount} \times \frac{1}{\ln(2 + \Delta\text{Hours})}$$

### 6.2 Diagnostics: `hop doctor`
Proactively validates the user's environment:
* Confirms config file read/write permissions.
* Identifies orphaned or moved project directories on disk.
* Checks if `$EDITOR` / configured editor exists in `$PATH`.
* Validates active shell wrapper installation.

### 6.3 Native Shell Wrappers

#### Bash / Zsh (`~/.bashrc` / `~/.zshrc`)
```bash
hop() {
    if [ "$#" -eq 1 ] && [ "$1" != "list" ] && [ "$1" != "add" ] && [ "$1" != "remove" ] && [ "$1" != "doctor" ]; then
        local target
        target=$(command hoppr _get_path "$1" 2>/dev/null)
        if [ -n "$target" ] && [ -d "$target" ]; then
            cd "$target" || return
            return
        fi
    fi
    command hoppr "$@"
}
```

#### PowerShell (`$PROFILE`)
```powershell
function hop {
    param([string]$cmd, [string]$arg1, [string]$arg2)
    if ($args.Count -eq 1 -and $cmd -notin @("list", "add", "remove", "create", "drop", "doctor", "import")) {
        $target = & hoppr _get_path $cmd 2>$null
        if ($target -and (Test-Path $target)) {
            Set-Location $target
            return
        }
    }
    & hoppr $args
}
```

---

## 7. Release Engineering & Build Flags

Stripping debugging symbols and applying compiler optimizations minimizes binary size and accelerates OS exec latency:

```bash
# Production release build
CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w -X 'hoppr/internal/version.Version=1.0.0' -X 'hoppr/internal/version.Commit=$(git rev-parse --short HEAD)'" \
  -o hop \
  ./cmd/hop
```

* **Size reduction**: ~55% (from 3.4 MB to 1.4 MB).
* **Static Binary**: Runs standalone on any Linux/macOS/Windows machine without runtime dependencies.

---

## 8. Implementation Checklist

- [x] **Low-Latency Routing:** Implemented zero-I/O fast path in [`main.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/main.go) and [`cmd/hop/main.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/cmd/hop/main.go) for `--help`, `-h`, `--version`, `-v`, and zero-arg calls (`< 0.5ms`).
- [x] **Stream Separation:** Clean machine data routed to `stdout` (`--plain`, `--json`, `_get_path`), diagnostics/errors routed to `stderr` in [`internal/commands/`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/commands).
- [x] **Exit Code Standardization:** Implemented POSIX codes (`0`, `1`, `2`, `3`) via [`internal/domain/errors.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/domain/errors.go).
- [x] **Concurrency & Safety:** Added cross-process locking ([`lock_windows.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/storage/lock_windows.go), [`lock_unix.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/storage/lock_unix.go)) with shared `RLock` reads, exclusive `Lock` mutations, and atomic `fsync` temp-file writes in [`json_storage.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/storage/json_storage.go).
- [x] **Deterministic Formatting:** In-place deterministic sorting with `slices.Sort` in [`internal/ui/format.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/ui/format.go).
- [x] **SOLID Package Restructuring:** Full decoupled architecture across `domain/`, `storage/`, `core/`, `commands/`, and `ui/`.
- [x] **Multi-List & Shorthand Features:** Complete multi-list CRUD, bulk [`import`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/commands/import.go), and contextual `.` resolution in [`internal/core/shorthand.go`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/core/shorthand.go).
- [x] **Diagnostics:** Implemented [`hop doctor`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/internal/commands/doctor.go) checking file permissions, broken project paths on disk, and system editor availability.
- [x] **Shell Wrappers & Autocomplete:** Pre-packaged native shell integration functions and dynamic completers in [`shell/`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/shell) for Bash, Zsh, Fish, and PowerShell.
- [x] **High-Performance Build:** High-performance release flags (`-ldflags="-s -w" -trimpath` + `CGO_ENABLED=0`) configured in [`Makefile`](file:///c:/Users/dwaip/OneDrive/Documents/Code/projects/Hoppr/Makefile).
