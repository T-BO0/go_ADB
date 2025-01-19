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

var recepiants = []string{"tornike.tabatadze@makingscience.com"}

var action Action

func main() {
	action = Action{
		Duration:   time.Duration(30 * time.Second),
		ThroughADB: true,
	}

	log.Println("starting")
	var reset bool
	defer func() {
		action.StopApp("ge.libertybank.business")
	}()
	go go_checkInternet()
	for {

		reset = false         // after restart reset restes to false so dont triger restart again and wait for 2 hours for it
		go go_restart(&reset) // gorutine that will trigers restart after 2 hours (savety - if app stockes this will make sure to automaticly restart)

		// checking internet if it is stabil continue execution else move to the start of the loop
		stability := action.CheckInternetStability()
		if stability >= 50 {
			fmt.Printf("internet stability is low: %f prc.\n going on retry!!", stability)
			continue
		}

		// unlock phone open app and auth
		startApp()

		// open options and click on tasks
		navigateToTaskPage()

		for {
			if reset {
				fmt.Println("restarting the bot")
				break
			}

			// if internet is not stable this will brake loop and move to parent loop wich will not start process until internet is stable
			stability := action.CheckInternetStability()
			fmt.Println("checking internet")
			if stability >= 50 {
				fmt.Printf("internet is not stable or lost the connection. \n Quiting, restrat once internet comes back")
				break
			}

			handlePayment() // if payment is visible this will confir otherwise refrashe the page by moving to tasks histor and back
		}
	}
}

// ANCHOR - handle payment
func handlePayment() {
	fmt.Println("checking if there is any payment visible!!")
	if !action.IsElementVisible("Keepz payment") {
		fmt.Println("there is no payment!")
		action.ClickByText("დავალებები")
		action.ClickByText("ავტორიზება")
	} else {
		fmt.Println("handling payment...")
		action.ClickByText("Keepz payment")
		action.ClickByText("ავტორიზება")
		for i := 0; i < 6; i++ {
			action.ClickByText("JKL")
		}
		action.ClickByText("მიმდინარე დავალება")
		action.ClickByText("ავტორიზება")
		fmt.Println("payment was handled successfully")
	}
}

// ANCHOR - start app
func startApp() {
	// connect to device if needed
	fmt.Println("connecting to device if needed")
	action.ConnectToDevice("192.168.1.13", "5555")
	// unlock screen if needed
	fmt.Println("unlocking screen if needed")
	action.UnlockScreen()
	// close app if it was open before
	fmt.Println("stoping the app if it is already running")
	action.StopApp("ge.libertybank.business")
	// start app
	fmt.Println("starting the app...")
	action.StartApp("ge.libertybank.business")
	// check if pin page is open
	fmt.Println("checking if pin page is open")
	if !action.IsElementVisible("JKL") {
		fmt.Println("failed to open app. could not open pin page")
		sendEmail("failed to open app. could not open pin page")
		panic("failed to open app. could not open pin page")
	}
	// insert pin
	fmt.Println("inserting pin...")
	for i := 0; i < 6; i++ {
		action.ClickByText("JKL")
	}

	fmt.Println("checking if main page was open")
	if !action.IsElementVisible("მთავარი") {
		fmt.Println("failed to authorize")
		sendEmail("failed to authorize")
		panic("failed to authorize. main page is not visible")
	}
	fmt.Println("successfully started the app")
	return
}

// ANCHOR - navigate to tasks
func navigateToTaskPage() {
	fmt.Println("navigating to services page")
	action.ClickByDescription("სერვისები")
	fmt.Println("navigating to payments page")
	action.ClickByText("მიმდინარე დავალება")

	fmt.Println("checking if payments page open")
	if !action.IsElementVisible("ჩემი ხელმოწერა") {
		fmt.Println("failed to open payments page")
		sendEmail("failed to open payments page, text 'ჩემი ხელმოწერა' was not visible")
		panic("failed to open payments page, text 'ჩემი ხელმოწერა' was not visible")
	}

	fmt.Println("payments page was opened successfully")
}

// go_rutine - Check internet connection
func go_checkInternet() {
	for {
		lostPackages := action.CheckInternetStability()
		if lostPackages > 50 && lostPackages <= 60 {
			sendEmail(fmt.Sprintf("internet is not stable we might lost connection! \n The bot will restart when internet is stable. \n instability level - %f", lostPackages))
		}
	}
}

// go_rutine - triger restart
func go_restart(resetable *bool) {
	time.Sleep(2 * time.Hour)
	*resetable = true
}

// Function to check battery level
func CheckBatteryLvl() int64 {
	output, err := RunAdbCommand(action.ThroughADB, "dumpsys", "battery", "|", "grep", "level")
	if err != nil {
		sendEmail(fmt.Sprintf("failed to check battery level. error: %v, cmd: 'dumpsys battery | grep level'", err))
		panic(fmt.Errorf("failed to check battery level. error: %v, cmd: 'dumpsys battery | grep level'", err))
	}
	return action.ExtractBatteryLvl(output)
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

// LINK -
type Action struct {
	Duration   time.Duration
	ThroughADB bool
}

// Function to check if element is visible based on text value
func (a *Action) IsElementVisible(text string) bool {
	var output string
	duration := a.Duration
	start := time.Now()
	for time.Since(start) < duration {
		RunAdbCommand(a.ThroughADB, "uiautomator", "dump")

		output, _ = RunAdbCommand(a.ThroughADB, "cat", "/sdcard/window_dump.xml")

		if strings.Contains(output, fmt.Sprintf("text=\"%s", text)) {
			return true
		}

		time.Sleep(1 * time.Second)
	}
	return false
}

// Function to click on an element based on text
func (a *Action) ClickByText(text string) error {
	if !a.IsElementVisible(text) {
		return fmt.Errorf("failed to locate element with text: %s within duration %d", text, a.Duration)
	}

	// read xml file for element
	doc, err := RunAdbCommand(a.ThroughADB, "cat", "/sdcard/window_dump.xml")
	if err != nil {
		panic(fmt.Errorf(`failed read content from 'window_dump.xml' for element with 
			text: '%s',
			error: %v,
			cmd: '%s'`,
			text, err, "cat /sdcard/window_dump.xml"))
	}

	// split xml content into arrays separated by "<node"
	arrdoc := strings.Split(doc, "<node")

	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("text=\"%s", text)) {
			num1, num2, err := a.calculateMiddlePoint(nod) // pass given nod will get bounds and caluclates middle point for element
			if err != nil {
				panic(fmt.Errorf(`failed read node bounds from 'window_dump.xml' for element with 
					text: '%s',
					error: %v,
					node: '%s'`,
					text, err, nod))
			}
			a.Click(num1, num2)
			break
		}
	}
	return nil
}

// Function to to Start app with given app package
func (a *Action) StartApp(appPackage string) error {
	_, err := RunAdbCommand(a.ThroughADB, "monkey", "-p", appPackage, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		return fmt.Errorf("failed to start app. package: '%s', error: %v", appPackage, err)
	}
	return nil
}

// Function to connect device over network
func (a *Action) ConnectToDevice(ip, port string) {
	// Define the ADB command and arguments
	adbCmd := "adb"
	connString := fmt.Sprintf("%s:%s", ip, port)
	args := []string{"connect", connString}

	// Execute the command
	cmd := exec.Command(adbCmd, args...)
	output, err := cmd.CombinedOutput() // Capture both stdout and stderr

	// Check for errors
	if err != nil {
		fmt.Printf("Error executing ADB command: %v\n", err)
		return
	}

	// Print the output
	fmt.Printf("ADB Output:\n%s\n", string(output))
}

// Function to Click on element based on Description
func (a *Action) ClickByDescription(desc string) error {

	if !a.IsElementWithhDescVisible(desc) {
		return fmt.Errorf("failed to locate lemenent with desc: %s within duration: %d", desc, a.Duration)
	}

	doc, err := RunAdbCommand(a.ThroughADB, "cat", "/sdcard/window_dump.xml")
	if err != nil {
		panic(fmt.Errorf(`failed read content from 'window_dump.xml' for element with 
			desc: '%s',
			error: %v,
			cmd: '%s'`,
			desc, err, "cat /sdcard/window_dump.xml"))
	}

	arrdoc := strings.Split(doc, "<node")
	for _, nod := range arrdoc {
		if strings.Contains(nod, fmt.Sprintf("content-desc=\"%s", desc)) {
			num1, num2, err := a.calculateMiddlePoint(nod)
			if err != nil {
				panic(fmt.Errorf(`failed read node bounds from 'window_dump.xml' for element with 
					desc: '%s',
					error: %v,
					node: '%s'`,
					desc, err, nod))
			}

			a.Click(num1, num2)
		}
	}
	return nil
}

// Function to check if element is visible based on desc. value
func (a *Action) IsElementWithhDescVisible(desc string) bool {
	var output string
	duration := a.Duration
	start := time.Now()

	for time.Since(start) < duration {
		RunAdbCommand(a.ThroughADB, "uiautomator", "dump")

		output, _ = RunAdbCommand(a.ThroughADB, "cat", "/sdcard/window_dump.xml")

		if strings.Contains(output, fmt.Sprintf("content-desc=\"%s", desc)) {
			return true
		}

		time.Sleep(1 * time.Second)
	}
	return false
}

// Function to stop an app with given package name
func (a *Action) StopApp(packageName string) {
	RunAdbCommand(a.ThroughADB, "am", "force-stop", packageName)
}

// Function to check if the specific window is focused
func (a *Action) IsFocusedOn(currentFocuse string) bool {
	duration := a.Duration
	start := time.Now()
	for time.Since(start) < duration {
		output, _ := RunAdbCommand(a.ThroughADB, "dumpsys", "window", "|", "grep", "mCurrentFocus")

		if strings.Contains(output, currentFocuse) {
			return true
		}

		time.Sleep(1 * time.Second)
	}
	return false
}

// Function to check if the screen is unlocked
func (a *Action) IsScreenUnlocked() bool {
	return a.IsFocusedOn("NotificationShade")
}

// Function to check if the screen is on NOTE: this does not check if the screen is locked or unlocked
// for checking if the screen locked or unlocked use IsScreenUnlocked()
func (a *Action) IsScreenOn() bool {
	start := time.Now()
	duration := a.Duration
	for time.Since(start) < duration {
		output, _ := RunAdbCommand(a.ThroughADB, "dumpsys", "display", "|", "grep", "mScreenState")

		if strings.Contains(output, "ON") {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}

// Function to unlock the screen
func (a *Action) UnlockScreen() {
	if !a.IsScreenOn() {
		// turn on the screen
		_, err := RunAdbCommand(a.ThroughADB, "input", "keyevent", "26")
		if err != nil {
			panic(fmt.Errorf("failed to unlock phone screen was on. err: %v, cmd:'%s'", err, "input keyevent 26"))
		}

		// unlock the screen
		_, err = RunAdbCommand(a.ThroughADB, "input", "keyevent", "82")
		if err != nil {
			panic(fmt.Errorf("failed to unlock phone screen. err: %v, cmd:'%s'", err, "input keyevent 82"))
		}
		return

	} else if !a.IsScreenUnlocked() {
		// unlock the screen
		_, err := RunAdbCommand(a.ThroughADB, "input", "keyevent", "82")
		if err != nil {
			panic(fmt.Errorf("failed to unlock phone screen!. err: %v, cmd:'%s'", err, "input keyevent 82"))
		}
	}
}

func (a *Action) Click(x, y int) error {
	_, err := RunAdbCommand(a.ThroughADB, "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	return err
}

// Function to extract bounds and calculate the middle point
func (a *Action) calculateMiddlePoint(xml string) (int, int, error) {
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

// Function to extract battery level
func (a *Action) ExtractBatteryLvl(output string) int64 {
	// Regular expression to extract digits
	re := regexp.MustCompile(`\d+`)

	// Find the first match (digits in the string)
	match := re.FindString(output)

	// Convert the string to an integer
	val, err := strconv.ParseInt(match, 10, 64)
	if err != nil {
		panic(fmt.Errorf("failed to extract battery level error: %v, from: %s", err, output))
	}
	return val
}

// Function will ping google to check internet stability
// It will return float32 between 0 <= t <= 100
// the higher the worst internet. 100 indicates no internet
func (a *Action) CheckInternetStability() float32 {
	output, err := RunAdbCommand(false, "ping", "-c", "10", "google.com")
	if err != nil {
		fmt.Println(err)
		return 100
	}
	res := a.extractPacketLoss(output)
	return res
}

func (a *Action) extractPacketLoss(pingOutput string) float32 {
	// Regex to capture the packet loss percentage
	re := regexp.MustCompile(`(\d+)% packet loss`)
	match := re.FindStringSubmatch(pingOutput)

	if len(match) < 2 {
		panic(fmt.Errorf("failed to capture packet loss percentage. output: %s", match))
	}

	// Convert the captured percentage to a float
	packetLoss, err := strconv.ParseInt(match[1], 10, 8)
	if err != nil {
		panic(fmt.Errorf("failed to parse packet loss percentage. error: %v", err))
	}

	return float32(packetLoss)
}

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
