package data

import (
	"strings"
	"wslc-tui-ms/internal/commands"
)

var AllCommands = map[string][]commands.Command{
	"Container": {
		{
			Name:        "ls",
			Full:        "wslc container ls --all --format json",
			Category:    "Container",
			Description: "List containers",
			Usage:       "wslc container ls [flags]",
			Examples: []string{
				"wslc container ls",
				"wslc container ls --all",
				"wslc container ls --all --format json",
				"wslc container ls --filter \"status=running\"",
			},
			Flags: []commands.Flag{
				{Short: "-a", Long: "--all", Description: "Show all containers (default shows just running)"},
				{Short: "", Long: "--format", Description: "Format the output using a Go template, e.g. '{{json .}}' or 'wide'", Default: "table"},
				{Short: "", Long: "--filter", Description: "Filter output based on conditions"},
				{Short: "", Long: "--no-trunc", Description: "Don't truncate output"},
			},
			Difficulty: "beginner",
			Tags:       []string{"list", "containers", "ps", "show", "view", "ls"},
		},
		{
			Name:        "run",
			Full:        "wslc run -d --name {name} {image}",
			Category:    "Container",
			Description: "Run a new container",
			Usage:       "wslc run [flags] IMAGE [COMMAND] [ARG...]",
			Examples: []string{
				"wslc run -d --name myapp ubuntu:latest",
				"wslc run -it --rm ubuntu:latest bash",
				"wslc run -d -p 8080:80 --name web nginx:latest",
				"wslc run --rm -it ubuntu:latest bash -c \"echo Hello world from WSL container!\"",
			},
			Flags: []commands.Flag{
				{Short: "-d", Long: "--detach", Description: "Run container in background"},
				{Short: "-i", Long: "--interactive", Description: "Keep STDIN open"},
				{Short: "-t", Long: "--tty", Description: "Allocate a pseudo-TTY"},
				{Short: "", Long: "--name", Description: "Assign a name to the container"},
				{Short: "-p", Long: "--publish", Description: "Publish a container's port(s) to the host"},
				{Short: "-v", Long: "--volume", Description: "Bind mount a volume"},
				{Short: "-e", Long: "--env", Description: "Set environment variables"},
				{Short: "", Long: "--rm", Description: "Automatically remove the container when it exits"},
				{Short: "", Long: "--network", Description: "Connect a container to a network"},
				{Short: "-w", Long: "--workdir", Description: "Working directory inside the container"},
				{Short: "", Long: "--cpus", Description: "Number of CPUs"},
				{Short: "-m", Long: "--memory", Description: "Memory limit"},
				{Short: "", Long: "--ulimit", Description: "Ulimit options"},
				{Short: "", Long: "--shm-size", Description: "Size of /dev/shm"},
				{Short: "", Long: "--stop-signal", Description: "Signal to stop the container"},
				{Short: "", Long: "--gpus", Description: "GPU devices (e.g. --gpus all)"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"run", "create", "start", "new", "launch"},
		},
		{
			Name:        "create",
			Full:        "wslc create --name {name} {image}",
			Category:    "Container",
			Description: "Create a new container without starting it",
			Usage:       "wslc create [flags] IMAGE [COMMAND] [ARG...]",
			Examples: []string{
				"wslc create --name myapp ubuntu:latest",
				"wslc create -p 8080:80 --name web nginx:latest",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--name", Description: "Assign a name to the container"},
				{Short: "-p", Long: "--publish", Description: "Publish a container's port(s) to the host"},
				{Short: "-v", Long: "--volume", Description: "Bind mount a volume"},
				{Short: "-e", Long: "--env", Description: "Set environment variables"},
				{Short: "", Long: "--network", Description: "Connect a container to a network"},
				{Short: "-w", Long: "--workdir", Description: "Working directory inside the container"},
				{Short: "", Long: "--cpus", Description: "Number of CPUs"},
				{Short: "-m", Long: "--memory", Description: "Memory limit"},
				{Short: "", Long: "--gpus", Description: "GPU devices"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"create", "new", "container"},
		},
		{
			Name:        "exec",
			Full:        "wslc exec -it {name} bash",
			Category:    "Container",
			Description: "Run a command in a running container",
			Usage:       "wslc exec [flags] CONTAINER COMMAND [ARG...]",
			Examples: []string{
				"wslc exec -it mycontainer bash",
				"wslc exec mycontainer ls -la /app",
				"wslc exec -u root mycontainer cat /etc/hosts",
			},
			Flags: []commands.Flag{
				{Short: "-d", Long: "--detach", Description: "Run command in background"},
				{Short: "-e", Long: "--env", Description: "Set environment variables"},
				{Short: "-i", Long: "--interactive", Description: "Keep STDIN open"},
				{Short: "", Long: "--privileged", Description: "Give extended privileges"},
				{Short: "-t", Long: "--tty", Description: "Allocate a pseudo-TTY"},
				{Short: "-u", Long: "--user", Description: "Username or UID"},
				{Short: "-w", Long: "--workdir", Description: "Working directory inside the container"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"exec", "run", "command", "shell", "bash", "interactive"},
		},
		{
			Name:        "logs",
			Full:        "wslc logs {name}",
			Category:    "Container",
			Description: "Fetch logs of a container",
			Usage:       "wslc logs [flags] CONTAINER",
			Examples: []string{
				"wslc logs mycontainer",
				"wslc logs -f mycontainer",
				"wslc logs --tail 100 mycontainer",
				"wslc logs --timestamps mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--follow", Description: "Follow log output"},
				{Short: "", Long: "--tail", Description: "Number of lines to show from the end"},
				{Short: "", Long: "--since", Description: "Show logs since timestamp"},
				{Short: "-t", Long: "--timestamps", Description: "Show timestamps"},
				{Short: "-n", Long: "--tail-n", Description: "Number of lines to show from the end"},
			},
			Difficulty: "beginner",
			Tags:       []string{"logs", "output", "debug", "follow", "tail"},
		},
		{
			Name:        "start",
			Full:        "wslc start {name}",
			Category:    "Container",
			Description: "Start one or more stopped containers",
			Usage:       "wslc start CONTAINER [CONTAINER...]",
			Examples: []string{
				"wslc start mycontainer",
				"wslc start container1 container2",
			},
			Flags: []commands.Flag{
				{Short: "-a", Long: "--attach", Description: "Attach STDOUT/STDERR and signal"},
				{Short: "-i", Long: "--interactive", Description: "Attach container's STDIN"},
			},
			Difficulty: "beginner",
			Tags:       []string{"start", "run", "boot", "resume"},
		},
		{
			Name:        "stop",
			Full:        "wslc stop {name}",
			Category:    "Container",
			Description: "Stop one or more running containers",
			Usage:       "wslc stop [flags] CONTAINER [CONTAINER...]",
			Examples: []string{
				"wslc stop mycontainer",
				"wslc stop -t 30 mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "-t", Long: "--time", Description: "Seconds to wait before killing", Default: "10"},
			},
			Difficulty: "beginner",
			Tags:       []string{"stop", "halt", "kill", "shutdown"},
		},
		{
			Name:        "kill",
			Full:        "wslc kill {name}",
			Category:    "Container",
			Description: "Kill one or more running containers",
			Usage:       "wslc kill [flags] CONTAINER [CONTAINER...]",
			Examples: []string{
				"wslc kill mycontainer",
				"wslc kill -s SIGTERM mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "-s", Long: "--signal", Description: "Signal to send to the container", Default: "KILL"},
			},
			Difficulty: "advanced",
			Tags:       []string{"kill", "stop", "force", "signal"},
		},
		{
			Name:        "rm",
			Full:        "wslc remove {name}",
			Category:    "Container",
			Description: "Remove one or more containers",
			Usage:       "wslc remove [flags] CONTAINER [CONTAINER...]",
			Examples: []string{
				"wslc remove mycontainer",
				"wslc remove -f mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Force removal of running container"},
			},
			Difficulty: "advanced",
			Tags:       []string{"remove", "delete", "rm", "cleanup"},
		},
		{
			Name:        "inspect",
			Full:        "wslc inspect {name}",
			Category:    "Container",
			Description: "Display detailed information on one or more containers",
			Usage:       "wslc inspect [flags] CONTAINER [CONTAINER...]",
			Examples: []string{
				"wslc inspect mycontainer",
				"wslc inspect --format '{{.State.Status}}' mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template"},
			},
			Difficulty: "beginner",
			Tags:       []string{"inspect", "info", "details", "json"},
		},
		{
			Name:        "stats",
			Full:        "wslc stats --format table",
			Category:    "Container",
			Description: "Display container resource usage statistics",
			Usage:       "wslc stats [flags] [CONTAINER...]",
			Examples: []string{
				"wslc stats",
				"wslc stats --format table",
				"wslc stats --all --format json",
			},
			Flags: []commands.Flag{
				{Short: "-a", Long: "--all", Description: "Show all containers regardless of state"},
				{Short: "", Long: "--format", Description: "Output formatting (json or table)", Default: "table"},
				{Short: "", Long: "--no-trunc", Description: "Do not truncate output"},
			},
			Difficulty: "beginner",
			Tags:       []string{"stats", "monitor", "cpu", "memory", "usage"},
		},
		{
			Name:        "attach",
			Full:        "wslc attach {name}",
			Category:    "Container",
			Description: "Attach stdin, stdout, and stderr to a running container",
			Usage:       "wslc attach [flags] CONTAINER",
			Examples: []string{
				"wslc attach mycontainer",
				"wslc attach --no-stdin mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--detach-keys", Description: "Override the default detach keys (default 'ctrl-p,ctrl-q')"},
				{Short: "", Long: "--no-stdin", Description: "Do not attach STDIN"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"attach", "interactive", "debug", "logs", "stdout", "stderr"},
		},
		{
			Name:        "export",
			Full:        "wslc export -o {file} {name}",
			Category:    "Container",
			Description: "Export a container's filesystem as a tar archive",
			Usage:       "wslc export [flags] CONTAINER",
			Examples: []string{
				"wslc export mycontainer -o backup.tar",
				"wslc export mycontainer > fs.tar",
			},
			Flags: []commands.Flag{
				{Short: "-o", Long: "--output", Description: "Write to a file, instead of STDOUT"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"export", "backup", "filesystem", "tar", "forensics", "debug"},
		},
		{
			Name:        "prune",
			Full:        "wslc container prune",
			Category:    "Container",
			Description: "Remove all stopped containers",
			Usage:       "wslc container prune [flags]",
			Examples: []string{
				"wslc container prune",
				"wslc container prune -f",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Do not prompt for confirmation"},
			},
			Difficulty: "advanced",
			Tags:       []string{"prune", "cleanup", "remove", "stopped"},
		},
	},
	"Image": {
		{
			Name:        "ls",
			Full:        "wslc image ls --format json",
			Category:    "Image",
			Description: "List images",
			Usage:       "wslc image ls [flags]",
			Examples: []string{
				"wslc image ls",
				"wslc image ls --all",
				"wslc image ls --format json",
			},
			Flags: []commands.Flag{
				{Short: "-a", Long: "--all", Description: "Show all images (not just dangling)"},
				{Short: "", Long: "--format", Description: "Format the output using a Go template, e.g. '{{json .}}' or 'wide'", Default: "table"},
				{Short: "", Long: "--filter", Description: "Filter output based on conditions"},
				{Short: "", Long: "--no-trunc", Description: "Don't truncate output"},
			},
			Difficulty: "beginner",
			Tags:       []string{"list", "images", "show", "ls"},
		},
		{
			Name:        "pull",
			Full:        "wslc pull {image}",
			Category:    "Image",
			Description: "Pull an image from a registry",
			Usage:       "wslc pull [flags] IMAGE",
			Examples: []string{
				"wslc pull ubuntu:latest",
				"wslc pull mcr.microsoft.com/dotnet/sdk:8.0",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--platform", Description: "Force image to a specific platform"},
			},
			Difficulty: "beginner",
			Tags:       []string{"pull", "download", "fetch", "registry"},
		},
		{
			Name:        "push",
			Full:        "wslc push {image}",
			Category:    "Image",
			Description: "Push an image or a repository to a registry",
			Usage:       "wslc push IMAGE[:TAG]",
			Examples: []string{
				"wslc push myapp:latest",
				"wslc push registry.example.com/myapp:latest",
			},
			Flags:      []commands.Flag{},
			Difficulty: "intermediate",
			Tags:       []string{"push", "upload", "registry"},
		},
		{
			Name:        "tag",
			Full:        "wslc tag {source} {target}",
			Category:    "Image",
			Description: "Create a new tag for an image",
			Usage:       "wslc tag SOURCE_IMAGE[:TAG] TARGET_IMAGE[:TAG]",
			Examples: []string{
				"wslc tag myapp:latest myapp:v1.0",
				"wslc tag myapp:latest registry.example.com/myapp:latest",
			},
			Flags:      []commands.Flag{},
			Difficulty: "beginner",
			Tags:       []string{"tag", "rename", "label", "registry"},
		},
		{
			Name:        "rm",
			Full:        "wslc rmi {image}",
			Category:    "Image",
			Description: "Remove one or more images",
			Usage:       "wslc rmi [flags] IMAGE [IMAGE...]",
			Examples: []string{
				"wslc rmi myapp:v1.0",
				"wslc rmi -f dangling-image-id",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Force removal"},
			},
			Difficulty: "advanced",
			Tags:       []string{"remove", "delete", "rmi", "cleanup"},
		},
		{
			Name:        "inspect",
			Full:        "wslc inspect {image}",
			Category:    "Image",
			Description: "Display detailed information on an image",
			Usage:       "wslc inspect [flags] IMAGE [IMAGE...]",
			Examples: []string{
				"wslc inspect ubuntu:latest",
				"wslc inspect --format '{{.Os}}/{{.Architecture}}' ubuntu:latest",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template"},
			},
			Difficulty: "beginner",
			Tags:       []string{"inspect", "info", "details", "json"},
		},
		{
			Name:        "prune",
			Full:        "wslc image prune",
			Category:    "Image",
			Description: "Remove unused images",
			Usage:       "wslc image prune [flags]",
			Examples: []string{
				"wslc image prune",
				"wslc image prune -f",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Do not prompt for confirmation"},
			},
			Difficulty: "advanced",
			Tags:       []string{"prune", "cleanup", "remove", "unused"},
		},
		{
			Name:        "save",
			Full:        "wslc save -o {file} {image}",
			Category:    "Image",
			Description: "Save one or more images to a tar archive",
			Usage:       "wslc save -o FILE IMAGE [IMAGE...]",
			Examples: []string{
				"wslc save -o ubuntu.tar ubuntu:latest",
			},
			Flags: []commands.Flag{
				{Short: "-o", Long: "--output", Description: "Write to a file, instead of STDOUT"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"save", "export", "tar", "backup"},
		},
		{
			Name:        "load",
			Full:        "wslc load -i {file}",
			Category:    "Image",
			Description: "Load an image from a tar archive or STDIN",
			Usage:       "wslc load [flags]",
			Examples: []string{
				"wslc load -i ubuntu.tar",
			},
			Flags: []commands.Flag{
				{Short: "-i", Long: "--input", Description: "Read from tar archive file, instead of STDIN"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"load", "import", "tar", "restore"},
		},
		{
			Name:        "build",
			Full:        "wslc build -t {tag} .",
			Category:    "Image",
			Description: "Build an image from a Containerfile/Dockerfile",
			Usage:       "wslc build [flags] PATH",
			Examples: []string{
				"wslc build -t myapp:latest .",
				"wslc build -f Dockerfile.prod -t myapp:prod .",
			},
			Flags: []commands.Flag{
				{Short: "-t", Long: "--tag", Description: "Name and optionally a tag for the image"},
				{Short: "-f", Long: "--file", Description: "Name of the Dockerfile/Containerfile"},
				{Short: "", Long: "--no-cache", Description: "Do not use cache when building the image"},
				{Short: "", Long: "--pull", Description: "Always attempt to pull a newer version"},
				{Short: "", Long: "--label", Description: "Set metadata for an image"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"build", "dockerfile", "containerfile", "image", "create"},
		},
		{
			Name:        "import",
			Full:        "wslc import {file} {image}",
			Category:    "Image",
			Description: "Import the contents from a tarball to create a filesystem image",
			Usage:       "wslc import file|URL|- [REPOSITORY[:TAG]] [flags]",
			Examples: []string{
				"wslc import backup.tar myapp:restored",
			},
			Flags: []commands.Flag{
				{Short: "-m", Long: "--message", Description: "Set commit message for imported image"},
				{Short: "", Long: "--platform", Description: "Set platform (e.g. linux/amd64)"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"import", "restore", "tar", "filesystem", "recovery"},
		},
	},
	"Network": {
		{
			Name:        "ls",
			Full:        "wslc network ls --format json",
			Category:    "Network",
			Description: "List networks",
			Usage:       "wslc network ls [flags]",
			Examples: []string{
				"wslc network ls",
				"wslc network ls --format json",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template, e.g. '{{json .}}' or 'wide'", Default: "table"},
				{Short: "", Long: "--filter", Description: "Provide filter values"},
			},
			Difficulty: "beginner",
			Tags:       []string{"list", "networks", "show", "ls"},
		},
		{
			Name:        "create",
			Full:        "wslc network create {name}",
			Category:    "Network",
			Description: "Create a new network",
			Usage:       "wslc network create [flags] NETWORK",
			Examples: []string{
				"wslc network create my-network",
				"wslc network create --driver bridge my-network",
				"wslc network create --subnet 172.20.0.0/16 my-network",
			},
			Flags: []commands.Flag{
				{Short: "-d", Long: "--driver", Description: "Driver to manage the Network", Default: "bridge"},
				{Short: "", Long: "--subnet", Description: "Subnet in CIDR format"},
				{Short: "", Long: "--gateway", Description: "Gateway IP address"},
				{Short: "", Long: "--internal", Description: "Restrict external access to the network"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"create", "new", "network", "bridge"},
		},
		{
			Name:        "rm",
			Full:        "wslc network rm {name}",
			Category:    "Network",
			Description: "Remove one or more networks",
			Usage:       "wslc network rm NETWORK [NETWORK...]",
			Examples: []string{
				"wslc network rm my-network",
			},
			Flags:      []commands.Flag{},
			Difficulty: "advanced",
			Tags:       []string{"remove", "delete", "rm", "network"},
		},
		{
			Name:        "inspect",
			Full:        "wslc network inspect {name}",
			Category:    "Network",
			Description: "Display detailed information on one or more networks",
			Usage:       "wslc network inspect [flags] NETWORK [NETWORK...]",
			Examples: []string{
				"wslc network inspect my-network",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template"},
			},
			Difficulty: "beginner",
			Tags:       []string{"inspect", "info", "details"},
		},
		{
			Name:        "connect",
			Full:        "wslc network connect {network} {container}",
			Category:    "Network",
			Description: "Connect a container to a network",
			Usage:       "wslc network connect [flags] NETWORK CONTAINER",
			Examples: []string{
				"wslc network connect my-network mycontainer",
				"wslc network connect --alias web my-network mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--alias", Description: "Add network alias for the container"},
				{Short: "", Long: "--ip", Description: "IPv4 address"},
				{Short: "", Long: "--link", Description: "Add link to another container"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"connect", "attach", "network"},
		},
		{
			Name:        "disconnect",
			Full:        "wslc network disconnect {network} {container}",
			Category:    "Network",
			Description: "Disconnect a container from a network",
			Usage:       "wslc network disconnect [flags] NETWORK CONTAINER",
			Examples: []string{
				"wslc network disconnect my-network mycontainer",
				"wslc network disconnect -f my-network mycontainer",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Force the container to disconnect from the network"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"disconnect", "detach", "network"},
		},
		{
			Name:        "prune",
			Full:        "wslc network prune",
			Category:    "Network",
			Description: "Remove all unused networks",
			Usage:       "wslc network prune [flags]",
			Examples: []string{
				"wslc network prune",
				"wslc network prune -f",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Do not prompt for confirmation"},
			},
			Difficulty: "advanced",
			Tags:       []string{"prune", "cleanup", "unused"},
		},
	},
	"Volume": {
		{
			Name:        "ls",
			Full:        "wslc volume ls --format json",
			Category:    "Volume",
			Description: "List volumes",
			Usage:       "wslc volume ls [flags]",
			Examples: []string{
				"wslc volume ls",
				"wslc volume ls --format json",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template, e.g. '{{json .}}' or 'wide'", Default: "table"},
				{Short: "", Long: "--filter", Description: "Provide filter values"},
			},
			Difficulty: "beginner",
			Tags:       []string{"list", "volumes", "show", "ls"},
		},
		{
			Name:        "create",
			Full:        "wslc volume create {name}",
			Category:    "Volume",
			Description: "Create a new volume",
			Usage:       "wslc volume create [flags] [VOLUME]",
			Examples: []string{
				"wslc volume create myvolume",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--label", Description: "Set metadata for a volume"},
				{Short: "", Long: "--driver", Description: "Specify volume driver (VHD-backed options: Uid, Gid, Fixed)"},
			},
			Difficulty: "beginner",
			Tags:       []string{"create", "new", "volume"},
		},
		{
			Name:        "rm",
			Full:        "wslc volume rm {name}",
			Category:    "Volume",
			Description: "Remove one or more volumes",
			Usage:       "wslc volume rm [flags] VOLUME [VOLUME...]",
			Examples: []string{
				"wslc volume rm myvolume",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Force the removal of a volume"},
			},
			Difficulty: "advanced",
			Tags:       []string{"remove", "delete", "rm", "volume"},
		},
		{
			Name:        "inspect",
			Full:        "wslc volume inspect {name}",
			Category:    "Volume",
			Description: "Display detailed information on one or more volumes",
			Usage:       "wslc volume inspect [flags] VOLUME [VOLUME...]",
			Examples: []string{
				"wslc volume inspect myvolume",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template"},
			},
			Difficulty: "beginner",
			Tags:       []string{"inspect", "info", "details"},
		},
		{
			Name:        "prune",
			Full:        "wslc volume prune",
			Category:    "Volume",
			Description: "Remove all unused local volumes",
			Usage:       "wslc volume prune [flags]",
			Examples: []string{
				"wslc volume prune",
				"wslc volume prune -f",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Do not prompt for confirmation"},
			},
			Difficulty: "advanced",
			Tags:       []string{"prune", "cleanup", "unused"},
		},
	},
	"Session": {
		{
			Name:        "ls",
			Full:        "wslc session ls",
			Category:    "Session",
			Description: "List sessions",
			Usage:       "wslc session ls [flags]",
			Examples: []string{
				"wslc session ls",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--format", Description: "Format the output using a Go template"},
			},
			Difficulty: "beginner",
			Tags:       []string{"list", "sessions", "show", "ls"},
		},
		{
			Name:        "enter",
			Full:        "wslc session enter {name}",
			Category:    "Session",
			Description: "Enter an existing session",
			Usage:       "wslc session enter SESSION",
			Examples: []string{
				"wslc session enter my-session",
			},
			Flags:      []commands.Flag{},
			Difficulty: "intermediate",
			Tags:       []string{"enter", "connect", "session"},
		},
		{
			Name:        "run",
			Full:        "wslc session run {name}",
			Category:    "Session",
			Description: "Run a command in a session (creates default session if needed)",
			Usage:       "wslc session run [flags] SESSION COMMAND [ARG...]",
			Examples: []string{
				"wslc session run my-session echo hello",
				"wslc session run --cpus 4 --memory 4096 my-session bash",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--cpus", Description: "Number of CPUs for the session"},
				{Short: "", Long: "--memory", Description: "Memory limit in MB for the session"},
				{Short: "", Long: "--storage", Description: "Storage location for session data"},
			},
			Difficulty: "intermediate",
			Tags:       []string{"run", "execute", "session"},
		},
		{
			Name:        "shell",
			Full:        "wslc session shell {name}",
			Category:    "Session",
			Description: "Open a shell in a session",
			Usage:       "wslc session shell [flags] SESSION",
			Examples: []string{
				"wslc session shell my-session",
			},
			Flags: []commands.Flag{
				{Short: "", Long: "--cpus", Description: "Number of CPUs for the session"},
				{Short: "", Long: "--memory", Description: "Memory limit in MB for the session"},
			},
			Difficulty: "beginner",
			Tags:       []string{"shell", "bash", "interactive", "session"},
		},
		{
			Name:        "terminate",
			Full:        "wslc session terminate {name}",
			Category:    "Session",
			Description: "Terminate a session and release resources",
			Usage:       "wslc session terminate SESSION",
			Examples: []string{
				"wslc session terminate my-session",
			},
			Flags:      []commands.Flag{},
			Difficulty: "advanced",
			Tags:       []string{"terminate", "stop", "destroy", "cleanup", "session"},
		},
	},
	"System": {
		{
			Name:        "info",
			Full:        "wslc info",
			Category:    "System",
			Description: "Display system-wide information",
			Usage:       "wslc info [flags]",
			Examples: []string{
				"wslc info",
			},
			Flags:      []commands.Flag{},
			Difficulty: "beginner",
			Tags:       []string{"info", "system", "details"},
		},
		{
			Name:        "prune",
			Full:        "wslc system prune",
			Category:    "System",
			Description: "Remove unused data (dangling images, stopped containers, unused networks)",
			Usage:       "wslc system prune [flags]",
			Examples: []string{
				"wslc system prune",
				"wslc system prune -f",
			},
			Flags: []commands.Flag{
				{Short: "-f", Long: "--force", Description: "Do not prompt for confirmation"},
				{Short: "", Long: "--volumes", Description: "Prune volumes too"},
			},
			Difficulty: "advanced",
			Tags:       []string{"cleanup", "prune", "remove", "unused", "disk"},
		},
		{
			Name:        "version",
			Full:        "wslc version",
			Category:    "System",
			Description: "Show the wslc version information",
			Usage:       "wslc version",
			Examples: []string{
				"wslc version",
			},
			Flags:      []commands.Flag{},
			Difficulty: "beginner",
			Tags:       []string{"version", "info", "debug", "compatibility"},
		},
	},
	"Registry": {
		{
			Name:        "login",
			Full:        "wslc login {name}",
			Category:    "Registry",
			Description: "Log in to a container registry",
			Usage:       "wslc login [flags] [SERVER]",
			Examples: []string{
				"wslc login",
				"wslc login -u myuser registry.example.com",
			},
			Flags: []commands.Flag{
				{Short: "-u", Long: "--username", Description: "Username"},
				{Short: "-p", Long: "--password", Description: "Password"},
				{Short: "", Long: "--password-stdin", Description: "Take the password from stdin"},
			},
			Difficulty: "beginner",
			Tags:       []string{"login", "auth", "registry", "credentials"},
		},
		{
			Name:        "logout",
			Full:        "wslc logout",
			Category:    "Registry",
			Description: "Log out from a container registry",
			Usage:       "wslc logout [flags] [SERVER]",
			Examples: []string{
				"wslc logout",
				"wslc logout registry.example.com",
			},
			Flags:      []commands.Flag{},
			Difficulty: "beginner",
			Tags:       []string{"logout", "auth", "registry"},
		},
	},
}

func GetCategories() []string {
	return []string{"Container", "Image", "Network", "Volume", "Session", "System", "Registry"}
}

func GetCommandsByCategory(category string) []commands.Command {
	return AllCommands[category]
}

func GetAllCommands() []commands.Command {
	var all []commands.Command
	for _, cmds := range AllCommands {
		all = append(all, cmds...)
	}
	return all
}

func init() {
	for category, catalog := range AllCommands {
		for index := range catalog {
			AllCommands[category][index].Schema = catalogSchema(category, catalog[index].Name)
		}
	}
}

func catalogSchema(category, name string) *commands.CommandSchema {
	key := category + "/" + name
	b := func(flag string) commands.Option {
		return commands.Option{Flag: flag, Kind: commands.OptionKindBoolean}
	}
	t := func(flag string) commands.Option {
		return commands.Option{Flag: flag, Kind: commands.OptionKindText}
	}
	td := func(flag, defaultValue string) commands.Option {
		return commands.Option{Flag: flag, Kind: commands.OptionKindText, Default: defaultValue}
	}
	s := func(flag string, choices ...string) commands.Option {
		return commands.Option{Flag: flag, Kind: commands.OptionKindSelect, Default: "table", Choices: choices}
	}
	n := func(flag, defaultValue string) commands.Option {
		return commands.Option{Flag: flag, Kind: commands.OptionKindNumeric, Default: defaultValue}
	}
	a := func(name string, required, repeatable bool) commands.Argument {
		return commands.Argument{Name: name, Required: required, Repeatable: repeatable}
	}

	schemas := map[string]*commands.CommandSchema{
		"Container/ls":      {Options: []commands.Option{b("--all"), td("--format", "table"), t("--filter"), b("--no-trunc")}},
		"Container/run":     {Arguments: []commands.Argument{a("image", true, false), a("command", false, true)}, Options: []commands.Option{b("--detach"), b("--interactive"), b("--tty"), t("--name"), t("--publish"), t("--volume"), t("--env"), b("--rm"), t("--network"), t("--workdir"), t("--cpus"), t("--memory"), t("--ulimit"), t("--shm-size"), t("--stop-signal"), t("--gpus")}},
		"Container/create":  {Arguments: []commands.Argument{a("image", true, false), a("command", false, true)}, Options: []commands.Option{t("--name"), t("--publish"), t("--volume"), t("--env"), t("--network"), t("--workdir"), t("--cpus"), t("--memory"), t("--gpus")}},
		"Container/exec":    {Arguments: []commands.Argument{a("container", true, false), a("command", true, true)}, Options: []commands.Option{b("--detach"), t("--env"), b("--interactive"), b("--privileged"), b("--tty"), t("--user"), t("--workdir")}},
		"Container/logs":    {Arguments: []commands.Argument{a("container", true, false)}, Options: []commands.Option{b("--follow"), t("--tail"), t("--since"), b("--timestamps"), t("--tail-n")}},
		"Container/start":   {Arguments: []commands.Argument{a("containers", true, true)}, Options: []commands.Option{b("--attach"), b("--interactive")}},
		"Container/stop":    {Arguments: []commands.Argument{a("containers", true, true)}, Options: []commands.Option{n("--time", "10")}},
		"Container/kill":    {Arguments: []commands.Argument{a("containers", true, true)}, Options: []commands.Option{td("--signal", "KILL")}},
		"Container/rm":      {Arguments: []commands.Argument{a("containers", true, true)}, Options: []commands.Option{b("--force")}},
		"Container/inspect": {Arguments: []commands.Argument{a("containers", true, true)}, Options: []commands.Option{t("--format")}},
		"Container/stats":   {Arguments: []commands.Argument{a("containers", false, true)}, Options: []commands.Option{s("--format", "table", "json"), b("--all"), b("--no-trunc")}},
		"Container/attach":  {Arguments: []commands.Argument{a("container", true, false)}, Options: []commands.Option{t("--detach-keys"), b("--no-stdin")}},
		"Container/export":  {Arguments: []commands.Argument{a("container", true, false)}, Options: []commands.Option{t("--output")}},
		"Container/prune":   {Options: []commands.Option{b("--force")}},

		"Image/ls":      {Options: []commands.Option{b("--all"), td("--format", "table"), t("--filter"), b("--no-trunc")}},
		"Image/pull":    {Arguments: []commands.Argument{a("image", true, false)}, Options: []commands.Option{t("--platform")}},
		"Image/push":    {Arguments: []commands.Argument{a("image", true, false)}},
		"Image/tag":     {Arguments: []commands.Argument{a("source", true, false), a("target", true, false)}},
		"Image/rm":      {Arguments: []commands.Argument{a("images", true, true)}, Options: []commands.Option{b("--force")}},
		"Image/inspect": {Arguments: []commands.Argument{a("images", true, true)}, Options: []commands.Option{t("--format")}},
		"Image/prune":   {Options: []commands.Option{b("--force")}},
		"Image/save":    {Arguments: []commands.Argument{a("images", true, true)}, Options: []commands.Option{t("--output")}},
		"Image/load":    {Options: []commands.Option{t("--input")}},
		"Image/build":   {Arguments: []commands.Argument{a("path", true, false)}, Options: []commands.Option{t("--tag"), t("--file"), b("--no-cache"), b("--pull"), t("--label")}},
		"Image/import":  {Arguments: []commands.Argument{a("file", true, false), a("image", false, false)}, Options: []commands.Option{t("--message"), t("--platform")}},

		"Network/ls":         {Options: []commands.Option{td("--format", "table"), t("--filter")}},
		"Network/create":     {Arguments: []commands.Argument{a("network", true, false)}, Options: []commands.Option{td("--driver", "bridge"), t("--subnet"), t("--gateway"), b("--internal")}},
		"Network/rm":         {Arguments: []commands.Argument{a("networks", true, true)}},
		"Network/inspect":    {Arguments: []commands.Argument{a("networks", true, true)}, Options: []commands.Option{t("--format")}},
		"Network/connect":    {Arguments: []commands.Argument{a("network", true, false), a("container", true, false)}, Options: []commands.Option{t("--alias"), t("--ip"), t("--link")}},
		"Network/disconnect": {Arguments: []commands.Argument{a("network", true, false), a("container", true, false)}, Options: []commands.Option{b("--force")}},
		"Network/prune":      {Options: []commands.Option{b("--force")}},

		"Volume/ls":      {Options: []commands.Option{td("--format", "table"), t("--filter")}},
		"Volume/create":  {Arguments: []commands.Argument{a("volume", false, false)}, Options: []commands.Option{t("--label"), t("--driver")}},
		"Volume/rm":      {Arguments: []commands.Argument{a("volumes", true, true)}, Options: []commands.Option{b("--force")}},
		"Volume/inspect": {Arguments: []commands.Argument{a("volumes", true, true)}, Options: []commands.Option{t("--format")}},
		"Volume/prune":   {Options: []commands.Option{b("--force")}},

		"Session/ls":        {Options: []commands.Option{t("--format")}},
		"Session/enter":     {Arguments: []commands.Argument{a("session", true, false)}},
		"Session/run":       {Arguments: []commands.Argument{a("session", true, false), a("command", true, true)}, Options: []commands.Option{t("--cpus"), t("--memory"), t("--storage")}},
		"Session/shell":     {Arguments: []commands.Argument{a("session", true, false)}, Options: []commands.Option{t("--cpus"), t("--memory")}},
		"Session/terminate": {Arguments: []commands.Argument{a("session", true, false)}},

		"System/info":     {},
		"System/prune":    {Options: []commands.Option{b("--force"), b("--volumes")}},
		"System/version":  {},
		"Registry/login":  {Arguments: []commands.Argument{a("server", false, false)}, Options: []commands.Option{t("--username"), t("--password"), b("--password-stdin")}},
		"Registry/logout": {Arguments: []commands.Argument{a("server", false, false)}},
	}
	schema := schemas[key]
	if schema == nil {
		return nil
	}
	return enrichCatalogSchema(schema, legacyCommand(category, name))
}

func legacyCommand(category, name string) commands.Command {
	for _, command := range AllCommands[category] {
		if command.Name == name {
			return command
		}
	}
	return commands.Command{}
}

func enrichCatalogSchema(schema *commands.CommandSchema, command commands.Command) *commands.CommandSchema {
	placeholderNames := positionalPlaceholders(command, schema)
	for index := range schema.Arguments {
		argument := &schema.Arguments[index]
		if argument.Label == "" {
			argument.Label = titleLabel(argument.Name)
		}
		if argument.Placeholder == "" {
			argument.Placeholder = argument.Name
			if index < len(placeholderNames) {
				argument.Placeholder = placeholderNames[index]
			}
		}
	}
	for index := range schema.Options {
		option := &schema.Options[index]
		if option.Description == "" {
			for _, flag := range command.Flags {
				if flag.Long == option.Flag {
					option.Description = flag.Description
					break
				}
			}
		}
		if option.Description == "" {
			option.Description = titleLabel(strings.TrimLeft(option.Flag, "-"))
		}
	}
	return schema
}

func titleLabel(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "-", " ")
	if value == "" {
		return "Value"
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func positionalPlaceholders(command commands.Command, schema *commands.CommandSchema) []string {
	if len(schema.Arguments) == 0 {
		return nil
	}
	optionKinds := make(map[string]commands.OptionKind, len(schema.Options))
	for _, option := range schema.Options {
		optionKinds[option.Flag] = option.Kind
	}
	for _, flag := range command.Flags {
		for _, name := range []string{flag.Short, flag.Long} {
			if name == "" {
				continue
			}
			for option, kind := range optionKinds {
				if option == flag.Long {
					optionKinds[name] = kind
				}
			}
		}
	}
	fields := strings.Fields(command.Full)
	var placeholders []string
	for index := 0; index < len(fields); index++ {
		field := fields[index]
		if kind, ok := optionKinds[field]; ok {
			if kind != commands.OptionKindBoolean && index+1 < len(fields) {
				index++
			}
			continue
		}
		if strings.HasPrefix(field, "{") && strings.HasSuffix(field, "}") {
			placeholders = append(placeholders, strings.Trim(field, "{}"))
		}
	}
	return placeholders
}
