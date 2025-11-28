package gettext

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	//"syscall"
)

func GetSelection(typee string) (string, error) {
	if typee == "0" {
		if runtime.GOOS == "linux" {
			out, err := exec.Command("xsel", "-o", "-b").Output()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(out)), nil

		} else if runtime.GOOS == "windows" {
			ps := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Clipboard]::GetText()`
			cmd := exec.Command("powershell", "-Command", ps)
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.Output()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(out)), nil

		}

	} else if typee == "1" {

		if runtime.GOOS == "linux" {
			out, err := exec.Command("xsel", "-o", "-p").Output()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(out)), nil
		} else if runtime.GOOS == "windows" {
			ps := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("^c"); Start-Sleep -Milliseconds 100; [System.Windows.Forms.Clipboard]::GetText()`
			cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("Error executing PowerShell: %v", err)
			}
			return strings.TrimSpace(string(out)), nil
		}

	}
	return "", fmt.Errorf("unsupported OS")
}
