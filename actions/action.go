package actions

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Function to to Start app with given app package
func StartApp(appPackage string) {
	_, err := RunAdbCommand("shell", "monkey", "-p", appPackage, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		log.Fatal(err)
	}
}

// Function to click on an element based on text
func ClickByText(text string) {
	_, err := RunAdbCommand("shell", "uiautomator", "dump", "/sdcard/window_dump.xml")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := RunAdbCommand("shell", "cat", "/sdcard/window_dump.xml")
	if err != nil {
		log.Fatal(err)
	}

	arrdoc := strings.Split(doc, "<node")
	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("text=\"%s", text)) {
			num1, num2, err := calculateMiddlePoint(nod)
			if err != nil {
				log.Fatal(err)
			}
			Click(num1, num2)
			break
		}
	}
}

func ClickByDescription(desc string) {
	_, err := RunAdbCommand("shell", "uiautomator", "dump", "/sdcard/window_dump.xml")
	if err != nil {
		log.Fatal(err)
	}
	doc, err := RunAdbCommand("shell", "cat", "/sdcard/window_dump.xml")
	if err != nil {
		log.Fatal(err)
	}

	arrdoc := strings.Split(doc, "<node")
	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("content-desc=\"%s\"", desc)) {
			num1, num2, err := calculateMiddlePoint(nod)
			if err != nil {
				log.Fatal(err)
			}
			Click(num1, num2)
		}
	}
}

// Function to run an ADB command.
// Take arguments comma separated.
// Example: RunAdbCommand("shell", "input", "tap", "100", "100")
func RunAdbCommand(args ...string) (string, error) {
	// Create the command object
	cmd := exec.Command("adb", args...)

	// Capture standard output and standard error
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error: %v, stderr: %s", err, stderr.String())
	}

	// Return the output
	result := strings.TrimSpace(stdout.String())
	return result, nil
}

func Click(x, y int) error {
	_, err := RunAdbCommand("shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

// Function to extract bounds and calculate the middle point
func calculateMiddlePoint(xml string) (int, int, error) {
	// Regular expression to extract the bounds from the XML string
	re := regexp.MustCompile(`bounds="\[(\d+),(\d+)\]\[(\d+),(\d+)\]"`)
	matches := re.FindStringSubmatch(xml)

	if len(matches) != 5 {
		return 0, 0, fmt.Errorf("unable to extract bounds from XML")
	}

	// Parse the bounds into integers
	x1, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing x1: %w", err)
	}
	y1, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing y1: %w", err)
	}
	x2, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing x2: %w", err)
	}
	y2, err := strconv.Atoi(matches[4])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing y2: %w", err)
	}

	// Calculate the middle point
	middleX := (x2-x1)/2 + x1
	middleY := (y2-y1)/2 + y1

	return middleX, middleY, nil
}

// Function to check if the specific window is focused
func IsFocusedOn(currentFocuse string) bool {
	output, err := RunAdbCommand("shell", "dumpsys", "window", "|", "grep", "mCurrentFocus")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Contains(output, currentFocuse)
}

// Function to check if the screen is unlocked
func IsScreenUnlocked() bool {
	return IsFocusedOn("NotificationShade")
}

// Function to check if the screen is on NOTE: this does not check if the screen is locked or unlocked
// for checking if the screen locked or unlocked use IsScreenUnlocked()
func IsScreenOn() bool {
	output, err := RunAdbCommand("shell", "dumpsys", "display", "|", "grep", "mScreenState")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Contains(output, "ON")
}

// Function to unlock the screen
func UnlockScreen() {
	if !IsScreenOn() {
		// turn on the screen
		_, err := RunAdbCommand("shell", "input", "keyevent", "26")
		if err != nil {
			log.Fatal(err)
		}

		// unlock the screen
		_, err = RunAdbCommand("shell", "input", "keyevent", "82")
		if err != nil {
			log.Fatal(err)
		}

		return

	} else if !IsScreenUnlocked() {
		// unlock the screen
		_, err := RunAdbCommand("shell", "input", "keyevent", "82")
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Function to stop an app with given package name
func StopApp(packageName string) {
	RunAdbCommand("shell", "am", "force-stop", packageName)
}

// Function to check if element is visible based on text value
func IsElementVisible(text string) bool {
	_, err := RunAdbCommand("shell", "uiautomator", "dump")
	if err != nil {
		log.Fatal(err)
	}

	output, err := RunAdbCommand("shell", "cat", "/sdcard/window_dump.xml")
	if err != nil {
		log.Fatal(err)
	}
	return strings.Contains(output, fmt.Sprintf("text=\"%s", text))
}

func CheckInternetStability() uint8 {
	output, err := RunAdbCommand("shell", "ping", "-c", "10", "google.com")
	if err != nil {
		return 100
	}
	res := extractPacketLoss(output)
	return res
}

func extractPacketLoss(pingOutput string) uint8 {
	// Regex to capture the packet loss percentage
	re := regexp.MustCompile(`(\d+)% packet loss`)
	match := re.FindStringSubmatch(pingOutput)

	if len(match) < 2 {
		log.Fatal("failed to capture packet loss percentage")
	}

	// Convert the captured percentage to a float
	packetLoss, err := strconv.ParseInt(match[1], 10, 8)
	if err != nil {
		log.Fatal("failed to parse packet loss percentage: ", err)
	}

	return uint8(packetLoss)
}

// Function to check battery level
func CheckBatteryLvl() int64 {
	output, err := RunAdbCommand("shell", "dumpsys", "battery", "|", "grep", "level")
	if err != nil {
		log.Fatal(err)
	}
	return extractBatteryLvl(output)
}

func extractBatteryLvl(output string) int64 {
	// Regular expression to extract digits
	re := regexp.MustCompile(`\d+`)

	// Find the first match (digits in the string)
	match := re.FindString(output)

	// Convert the string to an integer
	val, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		log.Fatal("failed to parse battery levle: ", val)
	}
	return val
}
