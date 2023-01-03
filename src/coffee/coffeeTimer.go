package coffee

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type coffeeTimer struct {
	armedLedGpio, disarmedLedGpio, armButtonGpio, checkStatusButtonGpio gpio.PinIO
	showStatusLengthMs, buttonPressLengthMs                             int
	isArmed                                                             bool
	triggerHour, triggerMin                                             int
	triggerFunc                                                         func()
	cancellableTimer                                                    *time.Timer
}

func NewCoffeeTimer(armedLedPin, disarmedLedPin, armButtonPin, checkStatusButtonPin int) *coffeeTimer {

	armedLedGpio := gpioreg.ByName(fmt.Sprint(armedLedPin))
	disarmedLedGpio := gpioreg.ByName(fmt.Sprint(disarmedLedPin))
	armButtonGpio := gpioreg.ByName(fmt.Sprint(armButtonPin))
	checkStatusButtonGpio := gpioreg.ByName(fmt.Sprint(checkStatusButtonPin))

	showStatusLengthMs := 2000
	buttonPressLengthMs := 300

	// Set it as input, with an internal pull down resistor:
	if err := checkStatusButtonGpio.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	if err := armButtonGpio.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	log.Printf("Check Status GPIO %s: %s\n", checkStatusButtonGpio, checkStatusButtonGpio.Function())
	log.Printf("Arm Timer GPIO %s: %s\n", armButtonGpio, armButtonGpio.Function())
	log.Printf("Armed LED GPIO %s: %s\n", armedLedGpio, armedLedGpio.Function())
	log.Printf("Diarmed LED GPIO %s: %s\n", disarmedLedGpio, disarmedLedGpio.Function())

	ct := coffeeTimer{armedLedGpio: armedLedGpio, disarmedLedGpio: disarmedLedGpio, armButtonGpio: armButtonGpio, checkStatusButtonGpio: checkStatusButtonGpio, showStatusLengthMs: showStatusLengthMs, buttonPressLengthMs: buttonPressLengthMs, isArmed: false}

	ct.SetTriggerTime(8, 30)
	ct.SetTriggerFunc(func() {})

	go func() {
		// Wait for edges as detected by the hardware, and print the value read:
		for {
			commenceWaiting := time.Now()

			// for some reason, the dge detection did not work unless we Read() the GPIO value
			checkStatusButtonGpio.Read()
			checkStatusButtonGpio.WaitForEdge(-1)
			if time.Since(commenceWaiting) > time.Duration(buttonPressLengthMs)*time.Millisecond {
				ct.showArmedStatus()
			}

			// for good measure, Read() again afterwards to be on the safe side..
			checkStatusButtonGpio.Read()
		}
	}()

	go func() {
		// Wait for edges as detected by the hardware, and print the value read:
		for {
			commenceWaiting := time.Now()

			armButtonGpio.Read()
			armButtonGpio.WaitForEdge(-1)
			if time.Since(commenceWaiting) > time.Duration(buttonPressLengthMs)*time.Millisecond {
				ct.toggleArmedStatus()
			}
			armButtonGpio.Read()
		}
	}()

	return &ct
}

func (ct coffeeTimer) showArmedStatus() {

	var statusGpio gpio.PinIO
	if ct.isArmed {
		statusGpio = ct.armedLedGpio
		log.Printf("CoffeeTimer Status: ARMED for %d:%02d\n", ct.triggerHour, ct.triggerMin)
	} else {
		statusGpio = ct.disarmedLedGpio
		log.Println("CoffeeTimer Status: disarmed")
	}

	//fmt.Println(statusGpio, "high")
	if err := statusGpio.Out(gpio.High); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(ct.buttonPressLengthMs) * time.Millisecond)

	//fmt.Println(statusGpio, "low")
	if err := statusGpio.Out(gpio.Low); err != nil {
		//fmt.Println("error gpio low")
		log.Fatal(err)
	}
}

func (ct *coffeeTimer) arm() {

	if ct.cancellableTimer != nil {
		// a timer is already going - stop it and create a new one below
		ct.disarm()
	}

	now := time.Now()

	triggerTime := time.Date(now.Year(), now.Month(), now.Day(), ct.triggerHour, ct.triggerMin, 0, 0, now.Location())
	if triggerTime.Before(now) {
		triggerTime = triggerTime.Add(24 * time.Hour)
	}
	ct.cancellableTimer = time.AfterFunc(time.Until(triggerTime), ct.triggerFunc)
	log.Println("CoffeeTimer triggering at", triggerTime)
	ct.isArmed = true

}

func (ct *coffeeTimer) disarm() {

	if ct.cancellableTimer != nil {
		// a timer is going - stop it
		ct.cancellableTimer.Stop()
	}

	ct.cancellableTimer = nil
	ct.isArmed = false

}

func (ct *coffeeTimer) toggleArmedStatus() {

	if ct.isArmed {
		ct.disarm()
	} else {
		ct.arm()
	}

	ct.showArmedStatus()

}

func (ct coffeeTimer) IsArmed() bool {
	return ct.isArmed
}

// sets the trigger for the next occurrence of HH:MM, usually tomorrow morning - DAYLIGHT SAVINGS BEHAVIOUR UNKNOWN!
func (ct *coffeeTimer) SetTriggerTime(hour, min int) {

	log.Printf("Setting trigger time to %d:%02d\n", hour, min)
	ct.triggerHour = hour
	ct.triggerMin = min
	log.Printf("Set trigger time to %d:%02d\n", ct.triggerHour, ct.triggerMin)

}

func (ct *coffeeTimer) SetTriggerFunc(f func()) {

	// ensure the timer gets disarmed after it has triggered the func
	funcWithDisarm := func() {
		log.Println("TRIGGERING!")
		f()
		ct.disarm()
	}

	ct.triggerFunc = funcWithDisarm

}

// sets both pins to High, as this is when the relay is turned off
func (ct coffeeTimer) Disconnect() {
	ct.armedLedGpio.Out(gpio.Low)
	ct.disarmedLedGpio.Out(gpio.Low)
}
