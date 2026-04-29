package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// Global mute state.
var isMuted = false

// Mute UI completely.
func MuteUI() {
	isMuted = true
}

//
// Functions that help print output prettier.
//

var (
	Cyan    = color.New(color.FgCyan).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Red     = color.New(color.FgRed).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
)

// Prints a green success message
func Success(msg string, args ...interface{}) {
	if isMuted {
		return
	}
	fmt.Printf("%s %s\n", Green("[✓]"), fmt.Sprintf(msg, args...))
}

// Prints a blue info message.
func Info(msg string, args ...interface{}) {
	if isMuted {
		return
	}
	fmt.Printf("%s %s\n", Cyan("[i]"), fmt.Sprintf(msg, args...))
}

// Prints a red error message.
func Error(msg string, args ...interface{}) {
	fmt.Printf("%s %s\n", Red("[×]"), fmt.Sprintf(msg, args...))
}

// Prints a yellow warning message.
func Warn(msg string, args ...interface{}) {
	if isMuted {
		return
	}
	fmt.Printf("%s %s\n", Yellow("[!]"), fmt.Sprintf(msg, args...))
}

// Prints an error message and then exits the program.
func Fatal(msg string, args ...interface{}) {
	Error(msg, args...)
	// You can add a little padding or a "Exiting..." message here
	os.Exit(1)
}

// Print a list item formatted.
func PrintListItem(w io.Writer, name string, status string, details string) {
	fmt.Fprintf(w, "  %s %s\t%s %s\n",
		Cyan("•"),
		name,
		Yellow("["+status+"]"),
		details,
	)
}

//
// Functions to help receive input.
//

// Internal helper to handle the "Ask" logic
func ask(defaultVal string, msg string, args ...interface{}) string {
	// Format the string here before printing!
	label := fmt.Sprintf(msg, args...)
	reader := bufio.NewReader(os.Stdin)

	if defaultVal != "" {
		fmt.Printf("%s %s [%s]: ", Magenta("[?]"), label, Cyan(defaultVal))
	} else {
		fmt.Printf("%s %s: ", Magenta("[?]"), label)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

// General text input prompt.
func Prompt(defaultVal string, msg string, args ...interface{}) string {
	return ask(defaultVal, msg, args...)
}

// Asks for a yes/no input from the user.
func Confirm(defaultY bool, msg string, args ...interface{}) bool {
	suffix := "y/N"
	if defaultY {
		suffix = "Y/n"
	}

	// Format the base message first
	label := fmt.Sprintf(msg, args...)

	// Append the suffix manually so it looks like: "Message (y/N)"
	promptMsg := fmt.Sprintf("%s (%s)", label, Cyan(suffix))

	// Pass "" as defaultVal to ask() so it doesn't print the [brackets] twice.
	res := strings.ToLower(ask("", "%s", promptMsg))

	// Now this correctly catches the empty Enter press!
	if res == "" {
		return defaultY
	}
	return res == "y" || res == "yes"
}

// Loading bars

// Wisp themed progress bar for downloads.
func NewDownloadBar(maxBytes int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		maxBytes,
		progressbar.OptionSetDescription(Cyan(description)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(25),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerCustom([]string{
			"·", "•", "●", "•", "·", " ", " ", " ",
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "~",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)
}

// Wisp themed progress bar for items
func NewItemBar(maxItems int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(
		maxItems,
		progressbar.OptionSetDescription(Cyan(description)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(25),
		progressbar.OptionThrottle(65),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "■",
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
	)
}
