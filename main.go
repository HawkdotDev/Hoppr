package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
)

type Config struct {
	Projects map[string]string `json:"projects"`
	Editor   string            `json:"editor"`
}

func getConfigPath() (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		os.Exit(1)
	}
	dir := filepath.Join(home, ".hoppr")
	return dir, filepath.Join(dir, "config.json")
}

func loadConfig() Config {
	_, file := getConfigPath()
	
	// Default config
	cfg := Config{
		Projects: make(map[string]string),
		Editor:   "code",
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return cfg // File doesn't exist, return default
	}

	json.Unmarshal(data, &cfg)
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]string)
	}
	return cfg
}

func saveConfig(cfg Config) {
	dir, file := getConfigPath()
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}

	os.WriteFile(file, data, 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hoppr <add|open|list|remove> [name]")
		return
	}

	cmd := os.Args[1]
	cfg := loadConfig()

	switch cmd {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a project name.")
			return
		}
		name := os.Args[2]
		cwd, _ := os.Getwd()
		cfg.Projects[name] = cwd
		saveConfig(cfg)
		fmt.Printf("Added project '%s' -> %s\n", name, cwd)

	case "list":
		if len(cfg.Projects) == 0 {
			fmt.Println("No projects saved yet.")
			return
		}
		// Use tabwriter for clean column alignment
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for name, path := range cfg.Projects {
			fmt.Fprintf(w, "%s\t: %s\n", name, path)
		}
		w.Flush()

	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a project name.")
			return
		}
		name := os.Args[2]
		if _, exists := cfg.Projects[name]; exists {
			delete(cfg.Projects, name)
			saveConfig(cfg)
			fmt.Printf("Removed project '%s'.\n", name)
		} else {
			fmt.Printf("Project '%s' not found.\n", name)
		}

	// HIDDEN ROUTINES FOR SHELL WRAPPER
	case "_get_path":
		if len(os.Args) >= 3 {
			fmt.Print(cfg.Projects[os.Args[2]])
		}
	case "_get_editor":
		fmt.Print(cfg.Editor)
	}
}