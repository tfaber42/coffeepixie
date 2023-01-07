package coffee

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

type RaspiConfig struct {
	EspressoButtonPin              int `yaml:"espresso_button_pin"`
	LungoButtonPin                 int `yaml:"lungo_button_pin"`
	ArmedLedPin                    int `yaml:"armed_led_pin"`
	DisarmedLedPin                 int `yaml:"disarmed_led_pin"`
	ArmButtonPin                   int `yaml:"arm_button_pin"`
	CheckStatusButtonPin           int `yaml:"check_status_button_pin"`
	ButtonPressDetectingDurationMs int `yaml:"button_press_detecting_duration_ms"`
}

var RaspiConfigDefaults = RaspiConfig{
	EspressoButtonPin:              27,
	LungoButtonPin:                 22,
	ArmedLedPin:                    17,
	DisarmedLedPin:                 4,
	ArmButtonPin:                   24,
	CheckStatusButtonPin:           23,
	ButtonPressDetectingDurationMs: 300,
}

var RaspiConfigNoInputButtons = RaspiConfig{
	EspressoButtonPin:              27,
	LungoButtonPin:                 22,
	ArmedLedPin:                    17,
	DisarmedLedPin:                 4,
	ArmButtonPin:                   -1,
	CheckStatusButtonPin:           -1,
	ButtonPressDetectingDurationMs: 300,
}

var NoRaspiInUseConfig = RaspiConfig{
	EspressoButtonPin:              -1,
	LungoButtonPin:                 -1,
	ArmedLedPin:                    -1,
	DisarmedLedPin:                 -1,
	ArmButtonPin:                   -1,
	CheckStatusButtonPin:           -1,
	ButtonPressDetectingDurationMs: 0,
}

type raspberrypi struct {
	espressoButtonGpio, lungoButtonGpio                                 gpio.PinIO
	armedLedGpio, disarmedLedGpio, armButtonGpio, checkStatusButtonGpio gpio.PinIO
	showArmedStatusFunc, toggleArmedStatusFunc                          func()
}

func NewRaspi(cfg RaspiConfig) raspberrypi {

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		fmt.Print("error init")
		log.Fatal(err)
	}

	espressoButtonGpio := getGPIOByNumber(cfg.EspressoButtonPin)
	lungoButtonGpio := getGPIOByNumber(cfg.LungoButtonPin)
	armedLedGpio := getGPIOByNumber(cfg.ArmedLedPin)
	disarmedLedGpio := getGPIOByNumber(cfg.DisarmedLedPin)
	armButtonGpio := getGPIOByNumber(cfg.ArmButtonPin)
	checkStatusButtonGpio := getGPIOByNumber(cfg.CheckStatusButtonPin)

	logGPIOFunction("Espresso button", espressoButtonGpio)
	logGPIOFunction("Lungo button", lungoButtonGpio)
	logGPIOFunction("Check Status button", checkStatusButtonGpio)
	logGPIOFunction("Arm Timer buttin", armButtonGpio)
	logGPIOFunction("Armed LED", armedLedGpio)
	logGPIOFunction("Diarmed LED", disarmedLedGpio)

	rp := raspberrypi{espressoButtonGpio: espressoButtonGpio, lungoButtonGpio: lungoButtonGpio, armedLedGpio: armedLedGpio, disarmedLedGpio: disarmedLedGpio, armButtonGpio: armButtonGpio, checkStatusButtonGpio: checkStatusButtonGpio}
	rp.SetShowArmedStatusFunc(func() {})
	rp.SetToggleArmedStatusFunc(func() {})

	// If configured, set button as input, with an internal pull down resistor, and start monitoring
	if checkStatusButtonGpio != nil {
		if err := checkStatusButtonGpio.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.Fatal(err)
		}

		go func() {
			// Wait for edges as detected by the hardware, and print the value read:
			for {
				commenceWaiting := time.Now()

				// for some reason, the dge detection did not work unless we Read() the GPIO value
				checkStatusButtonGpio.Read()
				checkStatusButtonGpio.WaitForEdge(-1)
				if time.Since(commenceWaiting) > time.Duration(cfg.ButtonPressDetectingDurationMs)*time.Millisecond {
					// TODO: test that this actually calls the new function if the function has been set from outside this NewRaspi function
					rp.showArmedStatusFunc()
				}

				// for good measure, Read() again afterwards to be on the safe side..
				checkStatusButtonGpio.Read()
			}
		}()
	}

	// If configured, set button as input, with an internal pull down resistor, and start monitoring
	if armButtonGpio != nil {
		if err := armButtonGpio.In(gpio.PullDown, gpio.RisingEdge); err != nil {
			log.Fatal(err)
		}

		go func() {
			// Wait for edges as detected by the hardware, and print the value read:
			for {
				commenceWaiting := time.Now()

				armButtonGpio.Read()
				armButtonGpio.WaitForEdge(-1)
				if time.Since(commenceWaiting) > time.Duration(cfg.ButtonPressDetectingDurationMs)*time.Millisecond {
					rp.toggleArmedStatusFunc()
				}
				armButtonGpio.Read()
			}
		}()
	}

	return rp
}

func (r raspberrypi) ActivateEspressoButton(press bool) {

	if r.espressoButtonGpio == nil {
		log.Println("Espresso button not configured for use, skipping setting to activate == ", press)
	}

	if press {
		log.Println("Pressing Espresso button")

		// Set the pin as output Low, as the relay is holding the button open when the pin is High
		if err := r.espressoButtonGpio.Out(gpio.Low); err != nil {
			log.Println("Error setting GPIO", r.espressoButtonGpio, ":", err)
		}
	} else {
		log.Println("Releasing Espresso button")

		// Set the pin as output High
		if err := r.espressoButtonGpio.Out(gpio.High); err != nil {
			log.Println("Error setting GPIO", r.espressoButtonGpio, ":", err)
		}
	}
}

func (r raspberrypi) ActivateLungoButton(press bool) {

	if r.lungoButtonGpio == nil {
		log.Println("Lungo button not configured for use, skipping setting to activate == ", press)
	}

	if press {
		log.Println("Pressing Lungo button")

		// Set the pin as output Low, as the relay is holding the button open when the pin is High
		if err := r.lungoButtonGpio.Out(gpio.Low); err != nil {
			log.Println("Error setting GPIO", r.espressoButtonGpio, " to Low:", err)
		}
	} else {
		log.Println("Releasing Lungo button")

		// Set the pin as output High
		if err := r.lungoButtonGpio.Out(gpio.High); err != nil {
			log.Println("Error setting GPIO", r.espressoButtonGpio, " to High:", err)
		}
	}
}

func (r raspberrypi) ActivateArmedStatusLED(isArmed bool, activateForMs int, logTriggerTime string) {

	var statusGpio gpio.PinIO
	if isArmed {
		statusGpio = r.armedLedGpio
		log.Printf("CoffeeTimer Status: ARMED for %s\n", logTriggerTime)
	} else {
		statusGpio = r.disarmedLedGpio
		log.Println("CoffeeTimer Status: disarmed")
	}

	if statusGpio == nil {
		log.Println("LED for status isArmed ==", isArmed, "is not configured for use, skipping activation")
	}

	if err := statusGpio.Out(gpio.High); err != nil {
		log.Println("Error setting GPIO", statusGpio, " to High:", err)
	}

	time.Sleep(time.Duration(activateForMs) * time.Millisecond)

	if err := statusGpio.Out(gpio.Low); err != nil {
		log.Println("Error setting GPIO", statusGpio, " to Low:", err)
	}
}

func (r *raspberrypi) SetShowArmedStatusFunc(f func()) {
	r.showArmedStatusFunc = f
}

func (r *raspberrypi) SetToggleArmedStatusFunc(f func()) {
	r.toggleArmedStatusFunc = f
}

func (r raspberrypi) Disconnect() {
	// sets both pins to High, as this is when the relay is turned off
	log.Println("Setting GPIO", r.espressoButtonGpio, "to High (which turns the Relay into Open status)")
	r.espressoButtonGpio.Out(gpio.High)
	log.Println("Setting GPIO", r.lungoButtonGpio, "to High (which turns the Relay into Open status)")
	r.lungoButtonGpio.Out(gpio.High)

	log.Println("Setting GPIO", r.armedLedGpio, "to Low")
	r.armedLedGpio.Out(gpio.Low)
	log.Println("Setting GPIO", r.disarmedLedGpio, "to Low")
	r.disarmedLedGpio.Out(gpio.Low)
}

func logGPIOFunction(descr string, g gpio.PinIO) {
	if g != nil {
		log.Printf("%s GPIO %s: %s\n", descr, g, g.Function())
	} else {
		log.Printf("%s GPIO not configured for use\n", descr)
	}
}

func getGPIOByNumber(n int) gpio.PinIO {
	if n < 0 {
		return nil
	}

	g := gpioreg.ByName(fmt.Sprint(n))

	if g == nil {
		log.Fatal("Failed to find Raspberry Pi GPIO", n)
	}

	return g
}
