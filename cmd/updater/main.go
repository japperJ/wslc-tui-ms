package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	if runtime.GOOS == "windows" && os.Getenv("WSLC_TUI_UPDATER_CHILD") != "1" {
		if err := handoffToTemp(*handoffPath); err != nil {
			fatal(err.Error())
		}
		return
	}
	logPath := filepath.Join(filepath.Dir(handoff.ResultPath), "update.log")
	logMessage(logPath, "handoff accepted: distribution=%s target=%s current=%s", handoff.Distribution, handoff.TargetVersion, handoff.CurrentExe)
	if err := update.WaitForProcessExit(handoff.ParentPID, 30*time.Second); err != nil {
		logMessage(logPath, "parent wait failed: %v", err)
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: err.Error()})
		fatal(err.Error())
	}
	if err := update.DownloadAndInstall(context.Background(), http.DefaultClient, handoff); err != nil {
		logMessage(logPath, "download/install failed: %v", err)
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: err.Error()})
		fatal(err.Error())
	}
	versionOutput, err := exec.Command(handoff.CurrentExe, "--version").CombinedOutput()
	logMessage(logPath, "installed executable version output: %s", strings.TrimSpace(string(versionOutput)))
	if err != nil || !strings.Contains(string(versionOutput), handoff.TargetVersion) {
		message := fmt.Sprintf("installed executable did not report %s", handoff.TargetVersion)
		if err != nil {
			message += ": " + err.Error()
		}
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: message})
		fatal(message)
	}
	if err := update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "installed"}); err != nil {
		fatal(err.Error())
	}
	command := exec.Command(handoff.CurrentExe, handoff.RelaunchArgs...)
	if handoff.WorkingDir != "" {
		command.Dir = handoff.WorkingDir
	}
	logMessage(logPath, "relaunching: %s %v (cwd=%s)", handoff.CurrentExe, handoff.RelaunchArgs, command.Dir)
	if err := command.Start(); err != nil {
		logMessage(logPath, "relaunch failed: %v", err)
		_ = update.WriteResult(handoff.ResultPath, update.Result{AttemptID: handoff.AttemptID, Version: handoff.TargetVersion, Status: "failed", Error: fmt.Sprintf("relaunch updated application: %v", err)})
		fatal(fmt.Sprintf("relaunch updated application: %v", err))
	}
	logMessage(logPath, "relaunch started with pid %d", command.Process.Pid)
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "wslc-tui updater:", message)
	os.Exit(1)
}

func logMessage(path, format string, args ...any) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339Nano), fmt.Sprintf(format, args...))
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func handoffToTemp(handoffPath string) error {
	current, err := os.Executable()
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(os.TempDir(), "wslc-tui-updater-*.exe")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := copyFile(current, temporaryPath); err != nil {
		os.Remove(temporaryPath)
		return err
	}
	command := exec.Command(temporaryPath, "-handoff", handoffPath)
	command.Env = append(os.Environ(), "WSLC_TUI_UPDATER_CHILD=1")
	if err := command.Start(); err != nil {
		os.Remove(temporaryPath)
		return err
	}
	return nil
}
