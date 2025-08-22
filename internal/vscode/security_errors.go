package vscode

import (
	"fmt"
	"strings"

	"github.com/DieGopherLT/vscode-terminal-runner/pkg/styles"
)

// handleSecureError processes and formats security-specific errors for better user experience.
// It analyzes the error message and provides contextual help and user-friendly messages.
func handleSecureError(err error) error {
	errMsg := err.Error()
	
	switch {
	case strings.Contains(errMsg, "authentication failed"):
		styles.PrintError("❌ Authentication failed. Bridge may have regenerated token.")
		styles.PrintInfo("Try restarting VSCode or the bridge extension.")
		return fmt.Errorf("authentication failed")
		
	case strings.Contains(errMsg, "rate limit exceeded"):
		styles.PrintError("❌ Too many requests. Wait before retrying.")
		styles.PrintInfo("Rate limit will reset in 5 minutes.")
		return fmt.Errorf("rate limit exceeded")
		
	case strings.Contains(errMsg, "command blocked"):
		styles.PrintError("❌ Command blocked by security policy.")
		styles.PrintInfo("The bridge has rejected this command as potentially unsafe.")
		return fmt.Errorf("command blocked by security policy")
		
	case strings.Contains(errMsg, "insecure permissions"):
		styles.PrintError("❌ Bridge file has insecure permissions.")
		styles.PrintInfo("Check file ownership and permissions (should be 0600 or 0700).")
		return fmt.Errorf("insecure file permissions")
		
	case strings.Contains(errMsg, "not in secure mode"):
		styles.PrintError("❌ Bridge is not running in secure mode.")
		styles.PrintInfo("Please enable secure mode in the VSCode extension settings.")
		return fmt.Errorf("bridge not in secure mode")
		
	case strings.Contains(errMsg, "bridge directory not found"):
		styles.PrintError("❌ Bridge directory not found.")
		styles.PrintInfo("Ensure the VSCode bridge extension is running and has created bridge files.")
		return fmt.Errorf("bridge directory not found")
		
	case strings.Contains(errMsg, "no valid secure bridge found"):
		styles.PrintError("❌ No valid secure bridge instances found.")
		styles.PrintInfo("Restart VSCode with the bridge extension enabled in secure mode.")
		return fmt.Errorf("no secure bridge found")
		
	case strings.Contains(errMsg, "invalid auth token"):
		styles.PrintError("❌ Invalid authentication token.")
		styles.PrintInfo("The bridge token may be corrupted or expired.")
		return fmt.Errorf("invalid authentication token")
		
	case strings.Contains(errMsg, "connection failed"):
		styles.PrintError("❌ Failed to connect to bridge.")
		styles.PrintInfo("Check that VSCode and the bridge extension are running.")
		return fmt.Errorf("connection failed")
		
	default:
		styles.PrintError(fmt.Sprintf("❌ Secure bridge error: %v", err))
		return err
	}
}