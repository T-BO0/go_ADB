package main

import (
	"log"
	"time"

	"github.com/T-BO0/go-ADB/actions"
)

func main() {
	log.Println("starting")
	c_network := make(chan string, 2)
	c_battery := make(chan string, 2)
	defer actions.StopApp("ge.libertybank.business")

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

		actions.UnlockScreen()
		actions.StartApp("ge.libertybank.business")
		for i := 0; i < 6; i++ {
			actions.ClickByText("JKL")
		}

		actions.ClickByDescription("სერვისები")

		actions.ClickByText("მიმდინარე დავალება")

		for {
			//SECTION - Check Internet Connection -->
			internet := <-c_network
			if internet == "internet is not available" {
				log.Println("internet is not available. Trying to reconnect")
				actions.StopApp("ge.libertybank.business")
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

			if !actions.IsElementVisible("Keepz payment") {
				actions.ClickByText("დავალებები")
				actions.ClickByText("ავტორიზება")
			} else {

				actions.ClickByText("Keepz payment")

				actions.ClickByText("ავტორიზება")

				for i := 0; i < 6; i++ {
					actions.ClickByText("JKL")
				}
				time.Sleep(3 * time.Second)

				actions.ClickByText("მიმდინარე დავალება")

				actions.ClickByText("ავტორიზება")
			}
		}
	}
}

// go_rutine - Check internet connection
func go_checkInternet(c_network chan<- string) {
	for {
		if actions.CheckInternetStability() <= 50 {
			c_network <- "internet is stable"
		} else if actions.CheckInternetStability() > 50 && actions.CheckInternetStability() <= 90 {
			c_network <- "internet is not stable"
		} else {
			c_network <- "internet is not available"
		}
	}
}

// go_rutine - Check battery level
func go_checkBattery(c_battery chan<- string) {
	for {
		if actions.CheckBatteryLvl() <= 25 {
			c_battery <- "battery is low"
		} else if actions.CheckBatteryLvl() > 25 && actions.CheckBatteryLvl() <= 60 {
			c_battery <- "battery is medium"
		} else {
			c_battery <- "battery is high"
		}
	}
}
