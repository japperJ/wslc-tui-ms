package main

import (
	"fmt"
	"os"
	"wslc-tui-ms/internal/app"
	"wslc-tui-ms/internal/buildinfo"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(buildinfo.Current())
		return
	}

	p := tea.NewProgram(app.NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
