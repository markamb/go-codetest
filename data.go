package main

import (
	"fmt"
	"os"
	"sync"
)

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

	mutex sync.Mutex // need to sync access as could have concurrent api calls
}

// PrintUpdate writes the current user data to the supplied File
func (d *Data) PrintUpdate(o *os.File, updateType string) {
	fmt.Fprintf(o, "User Data Updated: %s\n", updateType)
	fmt.Fprintf(o, "  WebsiteURL: %s\n", d.WebsiteURL)
	fmt.Fprintf(o, "  SessionID: %s\n", d.SessionID)
	fmt.Fprintf(o, "  ResizeFrom: (%d,%d)\n", d.ResizeFrom.Width, d.ResizeFrom.Height)
	fmt.Fprintf(o, "  ResizeTo: (%d,%d)\n", d.ResizeTo.Width, d.ResizeTo.Height)
	fmt.Fprintf(o, "  copyAndPaste controls:")
	for next := range d.CopyAndPaste {
		fmt.Fprintf(o, " %s", next)
	}
	fmt.Fprintf(o, "\n")
	fmt.Fprintf(o, "  FormCompletionTime: %d seconds\n", d.FormCompletionTime)
	fmt.Fprintf(o, "  websiteURLHashCode: %v\n", HashString(d.WebsiteURL))
}
