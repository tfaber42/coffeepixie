package coffee

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type CoffeeTimerConfig struct {
	TriggerTime string `yaml:"trigger_time"`
}

var CoffeeTimerConfigDefaults = CoffeeTimerConfig{
	TriggerTime: "8:30",
}

type coffeeTimer struct {
	raspi                               raspberrypi
	showStatusLengthMs                  int
	isArmed                             bool
	triggerHour, triggerMin, triggerSec int
	triggerFunc                         func()
	cancellableTimer                    *time.Timer
}

func NewCoffeeTimer(cfg CoffeeTimerConfig, raspi raspberrypi) *coffeeTimer {

	showStatusLengthMs := 2000

	ct := coffeeTimer{raspi: raspi, showStatusLengthMs: showStatusLengthMs, isArmed: false}

	ct.SetTriggerTime(cfg.TriggerTime)
	ct.SetTriggerFunc(func() {})

	// TODO: this only makes sense as a default while there's no (web) interface to explicitly arm the timer
	ct.arm()

	return &ct
}

func (ct *coffeeTimer) arm() {

	if ct.cancellableTimer != nil {
		// a timer is already going - stop it and create a new one below
		ct.disarm()
	}

	now := time.Now()

	triggerTime := time.Date(now.Year(), now.Month(), now.Day(), ct.triggerHour, ct.triggerMin, ct.triggerSec, 0, now.Location())
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

func (ct *coffeeTimer) ToggleArmedStatus() {

	if ct.isArmed {
		ct.disarm()
	} else {
		ct.arm()
	}

	ct.ShowArmedStatus()

}

func (ct coffeeTimer) ShowArmedStatus() {
	ct.raspi.ActivateArmedStatusLED(ct.isArmed, ct.showStatusLengthMs, fmt.Sprintf("%d:%02d", ct.triggerHour, ct.triggerMin))
}

func (ct coffeeTimer) IsArmed() bool {
	return ct.isArmed
}

// sets the trigger for the next occurrence of HH:MM, usually tomorrow morning - DAYLIGHT SAVINGS BEHAVIOUR UNKNOWN!
func (ct *coffeeTimer) SetTriggerTime(timeStr string) {

	var hour, min int
	var err error
	fields := strings.Split(timeStr, ":")
	if len(fields) < 2 || len(fields) > 3 {
		log.Printf("Unexpected trigger time format '%s', expected 'hh:mm[:ss]'\n", timeStr)
		log.Printf("Leaving trigger time unchanged at %d:%02d:%02d\n", ct.triggerHour, ct.triggerMin, ct.triggerSec)
		return
	}

	hour, err = strconv.Atoi(fields[0])
	if err != nil {
		log.Printf("Unexpected trigger time format '%s', expected 'hh:mm[:ss]'\n", timeStr)
		log.Printf("Leaving trigger time unchanged at %d:%02d:%02d\n", ct.triggerHour, ct.triggerMin, ct.triggerSec)
		return
	}

	min, err = strconv.Atoi(fields[1])
	if err != nil {
		log.Printf("Unexpected trigger time format '%s', expected 'hh:mm[:ss]'\n", timeStr)
		log.Printf("Leaving trigger time unchanged at %d:%02d:%02d\n", ct.triggerHour, ct.triggerMin, ct.triggerSec)
		return
	}

	sec := 0
	if len(fields) == 3 {
		sec, err = strconv.Atoi(fields[2])
		if err != nil {
			log.Printf("Unexpected trigger time format '%s', expected 'hh:mm:ss'\n", timeStr)
			log.Printf("Leaving trigger time unchanged at %d:%02d:%02d\n", ct.triggerHour, ct.triggerMin, ct.triggerSec)
			return
		}
	}

	log.Printf("Setting trigger time to %d:%02d:%02d\n", hour, min, sec)
	ct.triggerHour = hour
	ct.triggerMin = min
	ct.triggerSec = sec
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
