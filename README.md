# Hoppr

**The fast lane to your favourite projects.** Save, open, and switch between codebases from anywhere in your terminal.

## What is Hoppr?

`Hoppr` is a minimalist command-line tool that lets you save and open your favourite projects — from anywhere. No more digging through folders or copying paths. Just **hop** to your code.

Perfect for developers who work across multiple repos, microservices, or side projects and want speed without the clutter.

## Features

- Save any folder as a named project
- Instantly open projects with a single command
- List all saved projects
- Remove projects when you're done
- Works with your default terminal + editor setup

## Example Usage

```bash
# Add your current directory as a project named "myapp"
hoppr add myapp

# Hop into the "myapp" project
hoppr open myapp

# List all your saved projects
hoppr list

# Remove a saved project
hoppr remove myapp
```

## 📝 Configuration

By default, `hoppr` stores your project paths in:

```bash
~/.hoppr/config.json
```

You can customise:

* Default editor
* Terminal command
* Storage location (optional)

## Pro Tips

* Use `hoppr` with shell aliases or functions for even faster access.
* Combine with `fzf` or `tmux` for power setups.
* Works great in zsh, bash, fish, or any terminal.

## Roadmap

* [ ] Support for tags/groups
* [ ] Fuzzy search for project names
* [ ] Multi-editor support
* [ ] GUI launcher (maybe 👀)

## Contributing

Got ideas, bug reports, or feature requests?
Open an issue or submit a PR — contributions are welcome!

## Why Hoppr?

Because you **hop** between your favourite projects — fast, smooth, and effortless.
Just like a dev-powered rabbit.

## License

MIT License © \ HawkdotDev
