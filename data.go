package main

import "fmt"

//
// Maintains the UI usage data to be stored for a single user interaction
//

// Dimension represents a pages dimensions
type Dimension struct {
	Width  int
	Height int
}

// Data represents the data we want to capture from a users interaction with the page
type Data struct {
	WebsiteURL         string
	SessionID          string
	ResizeFrom         Dimension
	ResizeTo           Dimension
	CopyAndPaste       map[string]bool // map[fieldId]true
	FormCompletionTime int             // Seconds
}

// PrintUpdate writes the current user data to the screen after a user interaction has occured
func (d *Data) PrintUpdate(updateType string) {
	fmt.Printf("User Data Updated: %s\n", updateType)
	fmt.Printf("  WebsiteURL: %s\n", d.WebsiteURL)
	fmt.Printf("  SessionID: %s\n", d.SessionID)
	fmt.Printf("  ResizeFrom: (%d,%d)\n", d.ResizeFrom.Width, d.ResizeFrom.Height)
	fmt.Printf("  ResizeTo: (%d,%d)\n", d.ResizeTo.Width, d.ResizeTo.Height)
	fmt.Printf("  copyAndPaste controls:")
	for next := range d.CopyAndPaste {
		fmt.Printf(" %s", next)
	}
	fmt.Println()
	fmt.Printf("  FormCompletionTime: %d seconds\n", d.FormCompletionTime)
}
