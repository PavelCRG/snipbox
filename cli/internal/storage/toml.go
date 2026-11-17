package storage

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Command struct {
	Name        string   `toml:"name" json:"name"`
	Category    string   `toml:"category" json:"category"`
	Description string   `toml:"description" json:"description"`
	Tags        []string `toml:"tags" json:"tags"`
	Command     string   `toml:"command" json:"command"`
	Favorite    bool     `toml:"is_favorite" json:"is_favorite"`
	Usage       uint64   `toml:"usage_count" json:"usage_count"`
}
type file struct {
	Commands []Command `toml:"commands"`
}

func Path() string { h, _ := os.UserHomeDir(); return filepath.Join(h, ".snipbox", "commands.toml") }
func defaults() []Command {
	rows := [][5]string{
		{"Git", "Check working tree status", "status", "git status --short --branch", "git,working-tree"}, {"Git", "Stage all changes", "stage", "git add -A", "git,staging"}, {"Git", "Commit staged changes", "commit", "git commit -m \"{{message:Update files}}\"", "git,commit"}, {"Git", "Show recent commits", "log", "git log --oneline -{{count:10}}", "git,history"}, {"Git", "Create and switch to a branch", "branch", "git switch -c {{branch:feature/new-work}}", "git,branch"},
		{"Docker", "List running containers", "containers", "docker ps", "docker,containers"}, {"Docker", "List all containers", "all-containers", "docker ps -a", "docker,containers"}, {"Docker", "Show container logs", "logs", "docker logs -f {{container:app}}", "docker,logs"}, {"Docker", "Build an image", "build", "docker build -t {{image:my-app}} .", "docker,build"}, {"Docker", "Open a shell in a container", "shell", "docker exec -it {{container:app}} sh", "docker,shell"},
		{"System", "Show disk usage", "disk", "df -h", "system,disk"}, {"System", "Show memory usage", "memory", "free -h", "system,memory"}, {"System", "List processes", "processes", "ps aux", "system,processes"}, {"System", "Show current directory", "directory", "pwd", "system,files"}, {"System", "Find large files", "large-files", "find {{path:.}} -type f -size +{{size:100M}} -print", "system,files"},
		{"Network", "Show network interfaces", "interfaces", "ip addr", "network,interfaces"}, {"Network", "Test host connectivity", "ping", "ping -c {{count:4}} {{host:example.com}}", "network,diagnostics"}, {"Network", "Inspect DNS records", "dns", "dig {{host:example.com}}", "network,dns"}, {"Network", "Show listening ports", "ports", "ss -tulpn", "network,ports"}, {"Network", "Download a URL", "download", "curl -fL {{url:https://example.com}} -o {{file:download.bin}}", "network,http"},
		{"Tools", "Pretty-print JSON", "json", "jq . {{file:data.json}}", "tools,json"}, {"Tools", "Search text recursively", "search", "rg -n {{pattern:TODO}} {{path:.}}", "tools,search"}, {"Tools", "Compare files", "diff", "diff -u {{first:old.txt}} {{second:new.txt}}", "tools,files"}, {"Tools", "Show calendar", "calendar", "cal", "tools,date"}, {"Tools", "Create a SHA-256 checksum", "checksum", "sha256sum {{file:artifact.zip}}", "tools,security"},
	}
	commands := make([]Command, 0, 25)
	for _, r := range rows {
		commands = append(commands, Command{Name: r[0] + " · " + r[2], Category: r[0], Description: r[1], Command: r[3], Tags: strings.Split(r[4], ",")})
	}
	windows := map[int]string{
		10: "Get-PSDrive -PSProvider FileSystem", 11: "Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize,FreePhysicalMemory", 12: "Get-Process", 13: "Get-Location", 14: "Get-ChildItem {{path:.}} -File -Recurse | Where-Object Length -gt {{size:100MB}} | Select-Object FullName,Length",
		15: "ipconfig", 16: "ping -n {{count:4}} {{host:example.com}}", 17: "Resolve-DnsName {{host:example.com}}", 18: "Get-NetTCPConnection -State Listen", 19: "Invoke-WebRequest -Uri {{url:https://example.com}} -OutFile {{file:download.bin}}",
		20: "Get-Content {{file:data.json}} | ConvertFrom-Json | ConvertTo-Json -Depth 100", 21: "Get-ChildItem {{path:.}} -Recurse -File | Select-String -Pattern {{pattern:TODO}}", 22: "Compare-Object (Get-Content {{first:old.txt}}) (Get-Content {{second:new.txt}})", 23: "Get-Date", 24: "Get-FileHash -Algorithm SHA256 {{file:artifact.zip}}",
	}
	mac := map[int]string{11: "memory_pressure", 15: "ifconfig", 18: "lsof -i -P -n | grep LISTEN", 24: "shasum -a 256 {{file:artifact.zip}}"}
	if runtime.GOOS == "windows" {
		for i, command := range windows {
			commands[i].Command = command
		}
	} else if runtime.GOOS == "darwin" {
		for i, command := range mac {
			commands[i].Command = command
		}
	}
	return commands
}
func Load() ([]Command, error) {
	p := Path()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := Save(defaults()); err != nil {
			return nil, err
		}
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var f file
	if _, err = toml.Decode(string(b), &f); err != nil {
		return nil, err
	}
	return f.Commands, nil
}
func Save(commands []Command) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	err = toml.NewEncoder(f).Encode(file{commands})
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, p)
}
