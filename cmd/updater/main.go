package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
	"wslc-tui-ms/internal/update"
)

func main() {
	handoffPath := flag.String("handoff", "", "path to the update handoff JSON")
	flag.Parse()
	if *handoffPath == "" {
		fatal("-handoff is required")
	}
	handoff, err := update.ReadHandoff(*handoffPath)
	if err != nil {
		fatal(err.Error())
	}
	if err := update.WaitForProcessExit(handoff.ParentPID, 30*time.Second); err != nil {
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: err.Error()})
		fatal(err.Error())
	}
	if err := update.DownloadAndInstall(context.Background(), http.DefaultClient, handoff); err != nil {
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: err.Error()})
		fatal(err.Error())
	}
	if err := update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "installed"}); err != nil {
		fatal(err.Error())
	}
	command := exec.Command(handoff.CurrentExe, handoff.RelaunchArgs...)
	if handoff.WorkingDir != "" {
		command.Dir = handoff.WorkingDir
	}
	if err := command.Start(); err != nil {
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: fmt.Sprintf("relaunch updated application: %v", err)})
		fatal(fmt.Sprintf("relaunch updated application: %v", err))
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "wslc-tui updater:", message)
	os.Exit(1)
}
