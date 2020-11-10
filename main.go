package main

import (
	_ "github.com/Yesterday17/gd-suite/backends/all" // import all backends
	_ "github.com/Yesterday17/gd-suite/cmd/touch"
	"github.com/rclone/rclone/cmd"
	_ "github.com/rclone/rclone/cmd/all" // import all commands
	"github.com/rclone/rclone/fs"
	_ "github.com/rclone/rclone/lib/plugin" // import plugins
)

func main() {
	fs.Version = fs.Version + "-gd-suite-v1.0.0"
	cmd.Main()
}
