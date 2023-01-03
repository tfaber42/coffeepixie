package coffee

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type NespressoMachineConfig struct {
	EspressoButtonPin     int `yaml:"espresso_button_pin"`
	LungoButtonPin        int `yaml:"lungo_button_pin"`
	ButtonPressDurationMs int `yaml:"button_press_duration_ms"`
}

var NespressoMachineConfigDefaults = NespressoMachineConfig{
	EspressoButtonPin:     27,
	LungoButtonPin:        22,
	ButtonPressDurationMs: 300,
}

type nespressoMachine struct {
	espressoButtonGpio, lungoButtonGpio gpio.PinIO
	buttonPressLengthMs                 int
}

func NewNespressoMachine(cfg NespressoMachineConfig) nespressoMachine {

	espressoButtonGpio := gpioreg.ByName(fmt.Sprint(cfg.EspressoButtonPin))
	lungoButtonGpio := gpioreg.ByName(fmt.Sprint(cfg.LungoButtonPin))

	log.Printf("Espresso button GPIO %s: %s\n", espressoButtonGpio, espressoButtonGpio.Function())
	log.Printf("Lungo button GPIO %s: %s\n", lungoButtonGpio, lungoButtonGpio.Function())

	return nespressoMachine{espressoButtonGpio: espressoButtonGpio, lungoButtonGpio: lungoButtonGpio, buttonPressLengthMs: cfg.ButtonPressDurationMs}
}

func (n nespressoMachine) PressEspressoButton() {

	log.Println("Pressing Espresso button")

	// Set the pin as output Low.
	if err := n.espressoButtonGpio.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)

	// Set the pin as output High.
	if err := n.espressoButtonGpio.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
}

func (n nespressoMachine) PressLungoButton() {

	log.Println("Pressing Lungo button")

	// Set the pin as output Low.
	if err := n.lungoButtonGpio.Out(gpio.Low); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)

	// Set the pin as output High.
	if err := n.lungoButtonGpio.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
}

// sets both pins to High, as this is when the relay is turned off
func (n nespressoMachine) Disconnect() {
	n.espressoButtonGpio.Out(gpio.High)
	n.lungoButtonGpio.Out(gpio.High)
}
