# Hoppr

**The fast lane to your favourite projects.** Save, organise, and hop between codebases from anywhere in your terminal.

## What is Hoppr?

**Hoppr** is a minimalist CLI tool for developers who constantly context-switch. Save any directory as a named project, group projects into lists, and navigate your entire codebase landscape without ever touching a file picker. The command is `hop`.

No more `cd ../../..` archaeology. Just hop.

## Installation

```bash
go install github.com/hawkdotdev/hoppr@latest
```

Config is stored at `~/.hoppr/config.json` automatically on first use.

## Core Concepts

- **Project** — a named shortcut to a directory
- **List** — a named group of projects (e.g. `work`, `oss`, `clients`)
- **Default list** — the list used when no list is specified (starts as `default`)

## Commands

### Project Commands

```bash
hop add <name> [list]        # save current directory as a project
hop remove <name> [list]     # remove a project
```

### List Commands

```bash
hop create <list>            # create a new empty list
hop list [list]              # show all lists, or projects in a specific list
hop import <list> <folder>   # bulk-import a folder's subfolders as projects
hop drop <list>              # delete an entire list
hop rename <list> <new>      # rename a list
hop setdefault <list>        # set the default list
```

### Shorthand: `.`

Use `.` in place of any argument to resolve contextually:

| In place of | Resolves to |
|---|---|
| `<list>` | the default list |
| `<folder>` | current directory |
| `<name>` | current directory's basename |

## Example Workflow

```bash
# Create a dedicated list for work projects
hop create work

# In each project folder, add it to the list
cd ~/dev/api-service  &&  hop add api . work
cd ~/dev/frontend     &&  hop add . work      # "." uses folder name as project name

# Or bulk-import an entire workspace at once
hop import work ~/dev

# See what's in your list
hop list work
# [work]
#   api-service   : /home/you/dev/api-service
#   frontend      : /home/you/dev/frontend

# Make it the default so you can skip the list name
hop setdefault work

# Jump into a project (via shell function — see Shell Integration below)
hop frontend

# Clean up
hop remove api-service work
hop drop work
```

## Configuration

`~/.hoppr/config.json` — edited manually or managed by Hoppr:

```json
{
    "lists": {
        "default": {
            "myapp": "/home/you/dev/myapp"
        }
    },
    "editor": "code",
    "default_list": "default"
}
```

| Field | Description |
|---|---|
| `lists` | all your project lists |
| `editor` | editor opened when jumping to a project |
| `default_list` | list used when no list argument is given |

## Tips

- `hop import . .` — import all subfolders of the current directory into the default list
- Pair with `fzf` for fuzzy project switching
- Use separate lists for `work`, `oss`, `learning`, `clients` to keep things tidy
- `hop list` with no arguments gives you a full overview of everything saved

## License

MIT License © HawkdotDev
