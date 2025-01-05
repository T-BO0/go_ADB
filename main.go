package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.Println("starting")
	c_network := make(chan string, 2)
	c_battery := make(chan string, 2)
	defer StopApp("ge.libertybank.business")

	go go_checkInternet(c_network)
	go go_checkBattery(c_battery)
	for {
		//SECTION - Check Internet Connection -->
		network := <-c_network // from go_checkinternet
		if network == "internet is not available" {
			log.Println("internet is not available")
			continue
		} else if network == "internet is not stable" {
			log.Println("internet is not stable")
			continue
		}
		//!SECTION - Check Internet Connection <--

		//SECTION - Check Battery Level -->
		battery := <-c_battery // from go_checkBattery
		if battery == "battery is low" {
			log.Println("battery is low")
		} else if battery == "battery is medium" {
			log.Println("battery is medium")
		}
		//!SECTION - Check Battery Level <--

		UnlockScreen()
		StartApp("ge.libertybank.business")
		for i := 0; i < 6; i++ {
			ClickByText("JKL")
		}

		ClickByDescription("სერვისები")

		ClickByText("მიმდინარე დავალება")

		for {
			//SECTION - Check Internet Connection -->
			internet := <-c_network
			if internet == "internet is not available" {
				log.Println("internet is not available. Trying to reconnect")
				StopApp("ge.libertybank.business")
				break
			} else if network == "internet is not stable" {
				log.Println("internet is not stable, sending email")
				continue
			}
			//!SECTION - Check Internet Connection <--

			//SECTION - Check Battery Level -->
			battery := <-c_battery // from go_checkBattery
			if battery == "battery is low" {
				log.Println("battery is low, sending email")
			} else if battery == "battery is medium" {
				log.Println("battery is medium")
			}
			//!SECTION - Check Battery Level <--

			if !IsElementVisible("Keepz payment") {
				ClickByText("დავალებები")
				ClickByText("ავტორიზება")
			} else {

				ClickByText("Keepz payment")

				ClickByText("ავტორიზება")

				for i := 0; i < 6; i++ {
					ClickByText("JKL")
				}
				time.Sleep(3 * time.Second)

				ClickByText("მიმდინარე დავალება")

				ClickByText("ავტორიზება")
			}
		}
	}
}

// go_rutine - Check internet connection
func go_checkInternet(c_network chan<- string) {
	for {
		if CheckInternetStability() <= 50 {
			c_network <- "internet is stable"
		} else if CheckInternetStability() > 50 && CheckInternetStability() <= 90 {
			c_network <- "internet is not stable"
		} else {
			c_network <- "internet is not available"
		}
	}
}

// go_rutine - Check battery level
func go_checkBattery(c_battery chan<- string) {
	for {
		if CheckBatteryLvl() <= 25 {
			c_battery <- "battery is low"
		} else if CheckBatteryLvl() > 25 && CheckBatteryLvl() <= 60 {
			c_battery <- "battery is medium"
		} else {
			c_battery <- "battery is high"
		}
	}
}

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
