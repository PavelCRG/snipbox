package storage

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"strings"
)

type Command struct {
	Name        string   `toml:"name" json:"name"`
	Category    string   `toml:"category" json:"category"`
	Subcategory string   `toml:"subcategory,omitempty" json:"subcategory,omitempty"`
	Description string   `toml:"description" json:"description"`
	Tags        []string `toml:"tags" json:"tags"`
	Command     string   `toml:"command" json:"command"`
	Favorite    bool     `toml:"is_favorite" json:"is_favorite"`
	Usage       uint64   `toml:"usage_count" json:"usage_count"`
}

type file struct {
	Commands []Command `toml:"commands"`
}

func Path() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".snipbox", "commands.toml")
}

func defaults() []Command {
	return []Command{
		{Name: "Working tree status", Category: "Git", Subcategory: "Basics & Staging", Description: "Show the current branch and a concise summary of staged, unstaged, and untracked files. Run this before committing to understand exactly what Git will include.", Command: "git status --short --branch", Tags: []string{"git", "status"}},
		{Name: "Stage all changes", Category: "Git", Subcategory: "Basics & Staging", Description: "Stage every tracked modification, deletion, and new file in the repository. Review the staged diff afterward so unrelated files do not enter the same commit.", Command: "git add -A", Tags: []string{"git", "staging"}},
		{Name: "Review staged changes", Category: "Git", Subcategory: "Basics & Staging", Description: "Display the patch currently staged for the next commit. Use this to catch accidental edits, secrets, or generated files before they are recorded.", Command: "git diff --cached", Tags: []string{"git", "diff"}},
		{Name: "Review unstaged changes", Category: "Git", Subcategory: "Basics & Staging", Description: "Show edits that have not yet been staged. This helps separate work into focused commits and confirm the exact patch you intend to add.", Command: "git diff", Tags: []string{"git", "diff"}},
		{Name: "List local branches", Category: "Git", Subcategory: "Branches & Merging", Description: "List local branches and mark the branch currently checked out. Add `-a` when you also need to inspect remote tracking branches.", Command: "git branch --sort=-committerdate", Tags: []string{"git", "branch"}},
		{Name: "Create a feature branch", Category: "Git", Subcategory: "Branches & Merging", Description: "Create a branch from the current commit and switch to it immediately. Replace the suggested name with a short description of the work.", Command: "git switch -c {{branch:feature/new-work}}", Tags: []string{"git", "branch"}},
		{Name: "Switch branches", Category: "Git", Subcategory: "Branches & Merging", Description: "Switch to an existing local branch without changing its history. Git will stop if local edits would be overwritten, so commit or stash those edits first.", Command: "git switch {{branch:main}}", Tags: []string{"git", "branch"}},
		{Name: "Merge a branch", Category: "Git", Subcategory: "Branches & Merging", Description: "Merge the selected branch into the branch currently checked out and retain an explicit merge commit. Check the working tree first and resolve any conflicts Git reports.", Command: "git merge --no-ff {{branch:feature/new-work}}", Tags: []string{"git", "merge"}},
		{Name: "Recent commit graph", Category: "Git", Subcategory: "Branches & Merging", Description: "Show recent commits as a compact graph with branch and tag labels. Increase the count to trace a longer history when investigating how branches diverged.", Command: "git log --graph --oneline --decorate -{{count:20}}", Tags: []string{"git", "history"}},
		{Name: "Stash tracked changes", Category: "Git", Subcategory: "Undo & Stash", Description: "Save tracked working-tree changes in a named stash and restore a clean checkout. Add `-u` if new untracked files should be included too.", Command: "git stash push -m \"{{message:work in progress}}\"", Tags: []string{"git", "stash"}},
		{Name: "List stashes", Category: "Git", Subcategory: "Undo & Stash", Description: "List saved stashes with their identifiers and branch names. Use the identifier with `git stash show` or `git stash apply` to inspect or restore one.", Command: "git stash list", Tags: []string{"git", "stash"}},
		{Name: "Restore a file", Category: "Git", Subcategory: "Undo & Stash", Description: "Replace an unstaged file with the version from the index. This discards edits to that file, so inspect the diff first if you may need those changes.", Command: "git restore {{path:README.md}}", Tags: []string{"git", "restore"}},
		{Name: "Show configured remotes", Category: "Git", Subcategory: "Basics & Staging", Description: "List the remote names and fetch or push URLs configured for this repository. Check this before publishing if the project has multiple remotes or a fork.", Command: "git remote -v", Tags: []string{"git", "remote", "repository"}},
		{Name: "Fetch remote updates", Category: "Git", Subcategory: "Branches & Merging", Description: "Download new commits and branch references without changing the current branch or working files. Fetch before comparing your branch with its upstream or deciding whether it is safe to push.", Command: "git fetch --prune {{remote:origin}}", Tags: []string{"git", "fetch", "remote"}},
		{Name: "Publish the current branch", Category: "Git", Subcategory: "Branches & Merging", Description: "Push the current branch to the selected remote and set its upstream tracking branch. The upstream makes later `git pull` and `git push` commands shorter; verify the remote before publishing private work.", Command: "git push -u {{remote:origin}} HEAD", Tags: []string{"git", "push", "publish"}},
		{Name: "Compare branch changes", Category: "Git", Subcategory: "Branches & Merging", Description: "Show commits that are present on the current branch but absent from the base branch. This is a quick review of what a pull request will introduce; fetch first if the remote base may have moved.", Command: "git log --oneline {{base:origin/main}}..HEAD", Tags: []string{"git", "compare", "review"}},
		{Name: "Find a process", Category: "PowerShell", Subcategory: "Processes & Ports", Description: "Find running processes whose names match the supplied pattern. The output includes process IDs that can be used with other process commands.", Command: "Get-Process -Name {{name:code}}", Tags: []string{"powershell", "process"}},
		{Name: "Stop a process", Category: "PowerShell", Subcategory: "Processes & Ports", Description: "Stop a process by its process ID after confirming it is safe to close. Unsaved work in the process may be lost, and elevated permissions may be required.", Command: "Stop-Process -Id {{process_id:1234}} -Confirm", Tags: []string{"powershell", "process", "stop"}},
		{Name: "Inspect listening ports", Category: "PowerShell", Subcategory: "Processes & Ports", Description: "Show TCP connections in the listening state with their local addresses and process IDs. Use the process ID to identify which application owns a port.", Command: "Get-NetTCPConnection -State Listen | Sort-Object LocalPort", Tags: []string{"powershell", "network", "ports"}},
		{Name: "Operating system details", Category: "PowerShell", Subcategory: "System & Services", Description: "Read Windows version, build, and installation details from the operating system provider. This is useful when collecting system information for troubleshooting.", Command: "Get-CimInstance Win32_OperatingSystem | Select-Object Caption,Version,BuildNumber,LastBootUpTime", Tags: []string{"powershell", "system"}},
		{Name: "List running services", Category: "PowerShell", Subcategory: "System & Services", Description: "List Windows services that are currently running, sorted by display name. Add a name filter to narrow the output to a particular service family.", Command: "Get-Service | Where-Object Status -eq Running | Sort-Object DisplayName", Tags: []string{"powershell", "services"}},
		{Name: "Restart a service", Category: "PowerShell", Subcategory: "System & Services", Description: "Restart the named Windows service and wait for it to return to a running state. Administrative permissions may be required and active users of the service can be interrupted.", Command: "Restart-Service -Name {{service:Spooler}} -Force", Tags: []string{"powershell", "services"}},
		{Name: "List files recursively", Category: "PowerShell", Subcategory: "Files & Search", Description: "List files beneath a directory, including files in nested folders. Set the path variable to limit the scan to a specific project or location.", Command: "Get-ChildItem -Path {{path:.}} -File -Recurse", Tags: []string{"powershell", "files"}},
		{Name: "Search file contents", Category: "PowerShell", Subcategory: "Files & Search", Description: "Search text files recursively for a pattern and include line numbers in the results. Use a literal pattern when the input should not be interpreted as a regular expression.", Command: "Get-ChildItem {{path:.}} -File -Recurse | Select-String -Pattern '{{pattern:TODO}}'", Tags: []string{"powershell", "search"}},
		{Name: "Inspect file permissions", Category: "PowerShell", Subcategory: "Files & Search", Description: "Display the access control list for a file or directory to review which users and groups have permissions. Use this when investigating access failures or validating a security configuration.", Command: "Get-Acl -Path {{path:C:\\Projects\\app}} | Format-List", Tags: []string{"powershell", "files", "permissions"}},
		{Name: "Create a ZIP archive", Category: "PowerShell", Subcategory: "Files & Search", Description: "Package a file or directory into a ZIP archive, creating the destination archive if needed. Review the source and destination paths before running the command.", Command: "Compress-Archive -Path {{source:./output/*}} -DestinationPath {{archive:./output.zip}} -Force", Tags: []string{"powershell", "archive", "files"}},
		{Name: "Find the process using a port", Category: "PowerShell", Subcategory: "Processes & Ports", Description: "Map a listening TCP port to its owning process ID so you can identify a local server. Administrator permissions may be needed to see every process on the machine.", Command: "Get-NetTCPConnection -LocalPort {{port:8080}} -State Listen | Select-Object LocalAddress,LocalPort,OwningProcess", Tags: []string{"powershell", "ports", "process"}},
		{Name: "Inspect a process", Category: "PowerShell", Subcategory: "Processes & Ports", Description: "Show the executable path, start time, and ID for a running process. Use the ID to avoid ambiguity when several processes share a similar name.", Command: "Get-Process -Id {{process_id:1234}} | Select-Object Id,ProcessName,Path,StartTime", Tags: []string{"powershell", "process", "inspect"}},
		{Name: "Check service startup settings", Category: "PowerShell", Subcategory: "System & Services", Description: "Inspect a service's current state and startup mode without changing its configuration. This helps distinguish a stopped service from one configured not to start automatically.", Command: "Get-CimInstance Win32_Service -Filter \"Name='{{service:Spooler}}'\" | Select-Object Name,State,StartMode,StartName", Tags: []string{"powershell", "services", "system"}},
		{Name: "Find recently changed files", Category: "PowerShell", Subcategory: "Files & Search", Description: "List files modified within the chosen number of days and sort the newest first. Restrict the root path to a project directory to keep the search fast and relevant.", Command: "Get-ChildItem -Path {{path:.}} -File -Recurse | Where-Object LastWriteTime -gt (Get-Date).AddDays(-{{days:7}}) | Sort-Object LastWriteTime -Descending", Tags: []string{"powershell", "files", "search"}},
		{Name: "List running containers", Category: "Docker", Subcategory: "Containers", Description: "Show containers that are currently running with their image, status, and published ports. Add `-a` to include stopped containers.", Command: "docker ps", Tags: []string{"docker", "containers"}},
		{Name: "List all containers", Category: "Docker", Subcategory: "Containers", Description: "List every container on the host, including exited and created containers. This is useful for finding old containers that still need inspection or cleanup.", Command: "docker ps -a", Tags: []string{"docker", "containers"}},
		{Name: "Follow container logs", Category: "Docker", Subcategory: "Containers", Description: "Stream new log lines from a container and include a recent history for context. Press Ctrl+C to stop following the output.", Command: "docker logs --tail {{lines:100}} --follow {{container:app}}", Tags: []string{"docker", "logs"}},
		{Name: "Open a container shell", Category: "Docker", Subcategory: "Containers", Description: "Start an interactive shell inside a running container. Replace the container name and shell when the image does not include `sh`.", Command: "docker exec -it {{container:app}} sh", Tags: []string{"docker", "shell"}},
		{Name: "List local images", Category: "Docker", Subcategory: "Images", Description: "Show images stored on the local Docker host with their tags and sizes. Use this to verify a build or spot large images taking up disk space.", Command: "docker images", Tags: []string{"docker", "images"}},
		{Name: "Build an image", Category: "Docker", Subcategory: "Images", Description: "Build an image from the Dockerfile in a directory and assign it a useful local tag. Add `--no-cache` when you need to verify every build step from scratch.", Command: "docker build -t {{image:my-app}} {{path:.}}", Tags: []string{"docker", "build"}},
		{Name: "Inspect a container", Category: "Docker", Subcategory: "Containers", Description: "Print the container's configuration, network settings, mounts, and state as JSON. Pipe the result to a JSON query tool when you only need one field.", Command: "docker inspect {{container:app}}", Tags: []string{"docker", "inspect", "containers"}},
		{Name: "Show container resource usage", Category: "Docker", Subcategory: "Containers", Description: "Display live CPU, memory, network, and block I/O usage for running containers. Press Ctrl+C to stop the live refresh and return to the shell.", Command: "docker stats --no-stream {{container:app}}", Tags: []string{"docker", "stats", "performance"}},
		{Name: "Check Compose services", Category: "Docker", Subcategory: "Compose", Description: "Show the services declared by the Compose project, including their state and published ports. Run this from the directory containing the Compose file or pass its path explicitly.", Command: "docker compose ps", Tags: []string{"docker", "compose", "services"}},
		{Name: "Preview unused images", Category: "Docker", Subcategory: "Images", Description: "List image layers and build cache that Docker considers unused. Review the output before pruning because other stopped projects may rely on these local images.", Command: "docker system df", Tags: []string{"docker", "images", "disk"}},
		{Name: "Show disk space", Category: "Linux/Bash", Subcategory: "System & Services", Description: "Show available and used space for mounted filesystems in human-readable units. Use this when investigating a full disk or planning where to store large files.", Command: "df -h", Tags: []string{"linux", "disk"}},
		{Name: "Show memory usage", Category: "Linux/Bash", Subcategory: "System & Services", Description: "Display memory and swap totals in human-readable units. The available column is a more useful indicator of remaining memory than the free column alone.", Command: "free -h", Tags: []string{"linux", "memory"}},
		{Name: "List processes", Category: "Linux/Bash", Subcategory: "Processes & Ports", Description: "List processes with their owners, CPU and memory usage, and startup commands. Pipe the result to `less` on long-running hosts to browse it comfortably.", Command: "ps aux", Tags: []string{"linux", "process"}},
		{Name: "Show listening sockets", Category: "Linux/Bash", Subcategory: "Processes & Ports", Description: "List listening TCP and UDP sockets with process details where permissions allow. This helps identify services bound to unexpected ports.", Command: "ss -tulpn", Tags: []string{"linux", "ports"}},
		{Name: "Search project text", Category: "Linux/Bash", Subcategory: "Files & Search", Description: "Search files recursively for a text or regular-expression pattern and print matching line numbers. The path variable can restrict the search to one subtree.", Command: "rg -n '{{pattern:TODO}}' {{path:.}}", Tags: []string{"linux", "search"}},
		{Name: "Find large files", Category: "Linux/Bash", Subcategory: "Files & Search", Description: "Find files larger than a chosen size beneath a directory. Adjust the path and threshold to narrow a disk-usage investigation.", Command: "find {{path:.}} -type f -size +{{size:100M}} -print", Tags: []string{"linux", "files", "disk"}},
		{Name: "Measure directory sizes", Category: "Linux/Bash", Subcategory: "System & Services", Description: "Summarize the total size of each child directory in human-readable units. The depth limit keeps the output focused; increase it when you need a deeper breakdown.", Command: "du -h --max-depth={{depth:1}} {{path:.}} | sort -h", Tags: []string{"linux", "disk", "directories"}},
		{Name: "Find a command on PATH", Category: "Linux/Bash", Subcategory: "System & Services", Description: "Show which executable the shell will run for a command name. This is useful when multiple tool versions are installed or a command unexpectedly resolves to an alias.", Command: "command -v {{command:python}}", Tags: []string{"linux", "shell", "path"}},
		{Name: "Create a compressed archive", Category: "Linux/Bash", Subcategory: "Files & Search", Description: "Create a gzip-compressed tar archive from a directory while preserving its relative paths. Exclude cache or build directories explicitly when preparing a portable backup.", Command: "tar -czf {{archive:backup.tar.gz}} -C {{parent:.}} {{directory:project}}", Tags: []string{"linux", "archive", "backup"}},
		{Name: "Find files by name", Category: "Linux/Bash", Subcategory: "Files & Search", Description: "Search beneath a directory for file names matching a shell wildcard. The search is case-sensitive on most Linux filesystems; use a different pattern if capitalization varies.", Command: "find {{path:.}} -type f -iname '{{pattern:*.log}}' -print", Tags: []string{"linux", "find", "files"}},
		{Name: "Test host reachability", Category: "Network", Subcategory: "Connectivity", Description: "Send a small number of ICMP echo requests to a host and report packet loss and round-trip time. Some networks block ICMP even when the service itself is reachable.", Command: "ping -c {{count:4}} {{host:example.com}}", Tags: []string{"network", "ping"}},
		{Name: "Inspect DNS records", Category: "Network", Subcategory: "DNS", Description: "Query DNS records for a hostname and show the responding resolver and answer details. Use `A`, `AAAA`, or `MX` as the record type when needed.", Command: "dig {{host:example.com}} {{type:A}}", Tags: []string{"network", "dns"}},
		{Name: "Show local addresses", Category: "Network", Subcategory: "Connectivity", Description: "Display network interfaces and assigned addresses on Linux. Use this to identify the address and interface used by a local service.", Command: "ip addr", Tags: []string{"network", "interfaces"}},
		{Name: "Trace a network route", Category: "Network", Subcategory: "Connectivity", Description: "Trace the router hops between this machine and a destination host. Firewalls may hide intermediate hops, so missing responses do not always mean the destination is unreachable.", Command: "traceroute {{host:example.com}}", Tags: []string{"network", "route"}},
		{Name: "Check HTTP response headers", Category: "Network", Subcategory: "HTTP", Description: "Request only the response headers from a URL and print the HTTP status and metadata. This is a quick way to inspect redirects, caching, and server behavior.", Command: "curl -sSIL {{url:https://example.com}}", Tags: []string{"network", "http"}},
		{Name: "Test a TCP port", Category: "Network", Subcategory: "Connectivity", Description: "Attempt a TCP connection to a host and port with a short timeout. This distinguishes a reachable service from a DNS or routing failure.", Command: "curl -v --connect-timeout {{seconds:5}} telnet://{{host:example.com}}:{{port:443}}", Tags: []string{"network", "ports"}},
		{Name: "Check a TLS endpoint", Category: "Network", Subcategory: "TLS & Security", Description: "Open a TLS connection and print the server certificate chain and negotiated protocol. The command sends a clean input stream so it exits after the handshake instead of waiting for interactive input.", Command: "openssl s_client -connect {{host:example.com}}:{{port:443}} -servername {{server_name:example.com}} </dev/null", Tags: []string{"network", "tls", "certificate"}},
		{Name: "Trace DNS resolution", Category: "Network", Subcategory: "DNS", Description: "Query DNS iteratively from the root servers to the authoritative answer instead of relying on a recursive cache. This helps locate delegation or zone configuration problems.", Command: "dig +trace {{host:example.com}}", Tags: []string{"network", "dns", "troubleshooting"}},
		{Name: "Download a file with resume support", Category: "Network", Subcategory: "HTTP", Description: "Download a URL to a named local file and continue an existing partial download when the server supports byte ranges. Review the URL and destination before saving files from untrusted sources.", Command: "curl -fL --retry {{retries:3}} -C - -o {{file:download.bin}} {{url:https://example.com/file}}", Tags: []string{"network", "http", "download"}},
		{Name: "Inspect an HTTP response", Category: "Network", Subcategory: "HTTP", Description: "Request a URL and print the response status, headers, and timing details without saving the body. This makes it easier to diagnose redirects, slow connection phases, and server errors.", Command: "curl -sS -o /dev/null -w 'status=%{http_code} total=%{time_total}s\\n' {{url:https://example.com}}", Tags: []string{"network", "http", "diagnostics"}},
	}
}

func Load() ([]Command, error) {
	p := Path()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		commands := defaults()
		if err := Save(commands); err != nil {
			return nil, err
		}
		return commands, nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var f file
	if _, err = toml.Decode(string(b), &f); err != nil {
		return nil, err
	}
	commands, changed := migrate(f.Commands)
	if changed {
		if err := Save(commands); err != nil {
			return nil, err
		}
	}
	return commands, nil
}

func migrate(commands []Command) ([]Command, bool) {
	changed := false
	for i := range commands {
		switch commands[i].Category {
		case "System":
			commands[i].Category, changed = "Linux/Bash", true
			if commands[i].Subcategory == "" {
				commands[i].Subcategory = "System & Services"
			}
		case "Tools":
			commands[i].Category, changed = "Linux/Bash", true
			if commands[i].Subcategory == "" {
				commands[i].Subcategory = "Files & Search"
			}
		case "Git", "Docker", "Network", "PowerShell", "Linux/Bash":
		default:
			continue
		}
		if commands[i].Subcategory == "" {
			commands[i].Subcategory = "General"
			changed = true
		}
	}
	known := make(map[string]bool, len(commands))
	for _, command := range commands {
		known[command.Category+"\x00"+command.Name] = true
	}
	for _, command := range defaults() {
		key := command.Category + "\x00" + command.Name
		if !known[key] {
			commands = append(commands, command)
			known[key] = true
			changed = true
		}
	}
	return commands, changed
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

func Categories() []string {
	return []string{"Git", "PowerShell", "Docker", "Linux/Bash", "Network"}
}

func Subcategories(commands []Command, category string) []string {
	seen := map[string]bool{}
	var result []string
	for _, command := range commands {
		if command.Category == category && !seen[command.Subcategory] {
			seen[command.Subcategory] = true
			result = append(result, command.Subcategory)
		}
	}
	return result
}

func HasCategory(commands []Command, category string) bool {
	for _, command := range commands {
		if strings.EqualFold(command.Category, category) {
			return true
		}
	}
	return false
}
