package main

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const throughADB = true

var recepiants = []string{"tornike.tabatadze@makingscience.com"}

func min() {
	stable := CheckInternetStability()
	fmt.Printf("lost packages: %f\n", stable)
}

func main() {
	log.Println("starting")
	c_network := make(chan string)
	c_battery := make(chan string)
	defer func() {
		StopApp("ge.libertybank.business")
	}()

	go go_checkInternet(c_network)
	go go_checkBattery(c_battery)
	for {
		stability := CheckInternetStability()
		if stability >= 50 {
			fmt.Printf("internet stability is low: %f prc.\n going on retry!!", stability)
			continue
		}

		UnlockScreen()
		ConnectToDevice("192.168.1.18", "5555")
		StopApp("ge.libertybank.business")
		StartApp("ge.libertybank.business")
		for i := 0; i < 6; i++ {
			ClickByText("JKL")
		}
		ClickByDescription("სერვისები")
		ClickByText("მიმდინარე დავალება")

		for {
			stability := CheckInternetStability()
			fmt.Println("checking internet")
			if stability >= 50 {
				fmt.Printf("internet is not stable or lost the connection. \n Quiting, restrat once internet comes back")
				break
			}
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
		lostPackages := CheckInternetStability()
		if lostPackages <= 50 {
			c_network <- "internet is stable"
		} else if lostPackages > 50 && lostPackages <= 90 {
			c_network <- "internet is not stable"
			sendEmail(fmt.Sprintf("internet is not stable we might lost connection! \n The bot will restart when internet is stable. \n instability level - %f", lostPackages))
		} else {
			c_network <- "internet is not available"
			fmt.Println("no connection!!!")
		}
	}
}

// go_rutine - Check battery level
func go_checkBattery(c_battery chan<- string) {
	for {
		batteryLevel := CheckBatteryLvl()
		if batteryLevel <= 25 {
			c_battery <- "battery is low"
		} else if batteryLevel > 25 && batteryLevel <= 60 {
			c_battery <- "battery is medium"
			sendEmail(fmt.Sprintf("battery level is low! \n The bot will restart when internet is stable. \n Battery level - %d", batteryLevel))
		} else {
			c_battery <- "battery is high"
		}
	}
}

func ConnectToDevice(ip, port string) {
	connectionString := fmt.Sprintf("%s:%S", ip, port)
	_, err := RunAdbCommand(throughADB, "connect", connectionString)
	if err != nil {
		fmt.Printf("failed to connect to device with ip: %s, on port: %s ", ip, port)
		sendEmail(fmt.Sprintf("failed to connect to device with ip: %s, on port: %s ", ip, port))
	}
}

// Function to to Start app with given app package
func StartApp(appPackage string) error {
	_, err := RunAdbCommand(throughADB, "monkey", "-p", appPackage, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to start app. package: '%s', error: %v", appPackage, err)
	}
	return nil
}

// Function to click on an element based on text
func ClickByText(text string) error {
	_, err := RunAdbCommand(throughADB, "uiautomator", "dump", "/sdcard/window_dump.xml")
	if err != nil {
		log.Println(err)
		return fmt.Errorf(`failed dump window content to 'window_dump.xml' for element with 
			text: '%s',
			error: %v,
			cmd: '%s'`,
			text, err, "uiautomator dum /sdcard/window_dump.xml")
	}
	doc, err := RunAdbCommand(throughADB, "cat", "/sdcard/window_dump.xml")
	if err != nil {
		log.Println(err)
		return fmt.Errorf(`failed read content from 'window_dump.xml' for element with 
			text: '%s',
			error: %v,
			cmd: '%s'`,
			text, err, "cat /sdcard/window_dump.xml")
	}

	arrdoc := strings.Split(doc, "<node")
	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("text=\"%s", text)) {
			num1, num2, err := calculateMiddlePoint(nod)
			if err != nil {
				log.Println(err)
				return fmt.Errorf(`failed read node from 'window_dump.xml' for element with 
					text: '%s',
					error: %v,
					node: '%s'`,
					text, err, nod)
			}
			Click(num1, num2)
			break
		}
	}
	return nil
}

func ClickByDescription(desc string) error {
	_, err := RunAdbCommand(throughADB, "uiautomator", "dump", "/sdcard/window_dump.xml")
	if err != nil {
		log.Println(err)
		return fmt.Errorf(`failed dump window content to 'window_dump.xml' for element with 
			desc: '%s',
			error: %v,
			cmd: '%s'`,
			desc, err, "uiautomator dump /sdcard/window_dump.xml")
	}
	doc, err := RunAdbCommand(throughADB, "cat", "/sdcard/window_dump.xml")
	if err != nil {
		log.Println(err)
		return fmt.Errorf(`failed read content from 'window_dump.xml' for element with 
			desc: '%s',
			error: %v,
			cmd: '%s'`,
			desc, err, "cat /sdcard/window_dump.xml")
	}

	arrdoc := strings.Split(doc, "<node")
	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("content-desc=\"%s\"", desc)) {
			num1, num2, err := calculateMiddlePoint(nod)
			if err != nil {
				log.Println(err)
				return fmt.Errorf(`failed read node from 'window_dump.xml' for element with 
					desc: '%s',
					error: %v,
					node: '%s'`,
					desc, err, nod)
			}
			Click(num1, num2)
		}
	}
	return nil
}

// TODO - finish the rest
// Function to run an ADB command.
// Take arguments comma separated.
// Example: RunAdbCommand(throughADB, "input", "tap", "100", "100")
func RunAdbCommand(throughADB bool, args ...string) (string, error) {
	var cmd *exec.Cmd
	// Create the command object
	if throughADB {
		args = append([]string{"shell"}, args...)
		cmd = exec.Command("adb", args...)
		log.Println(cmd.String())
	} else {
		cmd = exec.Command("bash", args...) // need on mobile from pc it will not work (omit 'bash')
		log.Println(cmd.String())
	}
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
	_, err := RunAdbCommand(throughADB, "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
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
	output, err := RunAdbCommand(throughADB, "dumpsys", "window", "|", "grep", "mCurrentFocus")
	if err != nil {
		sendEmail(fmt.Sprintf(`failed check if focused on 
			content: '%s'
			error: %v`,
			currentFocuse, err))
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
	output, err := RunAdbCommand(throughADB, "dumpsys", "display", "|", "grep", "mScreenState")
	if err != nil {
		sendEmail(fmt.Sprintf("failed to check if phone screen. err: %v, cmd:'%s'", err, "dumpsys display | grep mScreenState"))
		log.Fatal(err)
	}
	return strings.Contains(output, "ON")
}

// Function to unlock the screen
func UnlockScreen() {
	if !IsScreenOn() {
		// turn on the screen
		_, err := RunAdbCommand(throughADB, "input", "keyevent", "26")
		if err != nil {
			sendEmail(fmt.Sprintf("failed to unlock phone screen was on. err: %v, cmd:'%s'", err, "input keyevent 26"))
			log.Fatal(err)
		}

		// unlock the screen
		_, err = RunAdbCommand(throughADB, "input", "keyevent", "82")
		if err != nil {
			sendEmail(fmt.Sprintf("failed to unlock phone screen. err: %v, cmd:'%s'", err, "input keyevent 82"))
			log.Fatal(err)
		}

		return

	} else if !IsScreenUnlocked() {
		// unlock the screen
		_, err := RunAdbCommand(throughADB, "input", "keyevent", "82")
		if err != nil {
			sendEmail(fmt.Sprintf("failed to unlock phone screen!. err: %v, cmd:'%s'", err, "input keyevent 82"))
			log.Fatal(err)
		}
	}
}

// Function to stop an app with given package name
func StopApp(packageName string) {
	RunAdbCommand(throughADB, "am", "force-stop", packageName)
}

// Function to check if element is visible based on text value
func IsElementVisible(text string) bool {
	_, err := RunAdbCommand(throughADB, "uiautomator", "dump")
	if err != nil {
		sendEmail(fmt.Sprintf("failed dump window content to 'window_dump.xml' for element with text: '%s', error: %v", text, err))
		log.Fatal(err)
	}

	output, err := RunAdbCommand(throughADB, "cat", "/sdcard/window_dump.xml")
	if err != nil {
		sendEmail(fmt.Sprintf(`failed to read content from 'window_dump.xml' and check element with 
			text: '%s', 
			error: %v, 
			cmd: cat /sdcard/window_dump.xml'`, text, err))
		log.Fatal(err)
	}
	return strings.Contains(output, fmt.Sprintf("text=\"%s", text))
}

func CheckInternetStability() float32 {
	output, err := RunAdbCommand(false, "ping", "-c", "10", "google.com")
	if err != nil {
		fmt.Println(err)
		return 100
	}
	res := extractPacketLoss(output)
	return res
}

func extractPacketLoss(pingOutput string) float32 {
	// Regex to capture the packet loss percentage
	re := regexp.MustCompile(`(\d+)% packet loss`)
	match := re.FindStringSubmatch(pingOutput)

	if len(match) < 2 {
		sendEmail(fmt.Sprintf("failed to capture packet loss percentage. output: %s", match))
		log.Fatal("failed to capture packet loss percentage")
	}

	// Convert the captured percentage to a float
	packetLoss, err := strconv.ParseInt(match[1], 10, 8)
	if err != nil {
		sendEmail(fmt.Sprintf("failed to parse packet loss percentage. error: %v", err))
		log.Fatal("failed to parse packet loss percentage: ", err)
	}

	return float32(packetLoss)
}

// Function to check battery level
func CheckBatteryLvl() int64 {
	output, err := RunAdbCommand(throughADB, "dumpsys", "battery", "|", "grep", "level")
	if err != nil {
		sendEmail(fmt.Sprintf("failed to check battery level. error: %v, cmd: 'dumpsys battery | grep level'", err))
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
		sendEmail(fmt.Sprintf("failed to extract battery level error: %v, from: %s", err, output))
		log.Fatal("failed to parse battery levle from: ", output)
	}
	return val
}

func sendEmail(text string) {
	// Sender email credentials
	from := "keepzbot@gmail.com"
	password := "agzo gpoi wyqn cbjo"

	// Recipient email addresses
	to := recepiants

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message body
	subject := "Subject: Liberty bot (payment)\n"
	body := text
	message := subject + "\n" + body

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(message))
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Email sent successfully!")
}
