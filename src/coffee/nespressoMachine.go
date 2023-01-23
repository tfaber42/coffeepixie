package coffee

import (
	"time"
)

type NespressoMachineConfig struct {
	ButtonPressDurationMs int `yaml:"button_press_duration_ms"`
}

var NespressoMachineConfigDefaults = NespressoMachineConfig{
	ButtonPressDurationMs: 300,
}

type NespressoMachine struct {
	raspi               raspberrypi
	buttonPressLengthMs int
}

func NewNespressoMachine(cfg NespressoMachineConfig, raspi raspberrypi) NespressoMachine {
	return NespressoMachine{raspi: raspi, buttonPressLengthMs: cfg.ButtonPressDurationMs}
}

func (n NespressoMachine) pressEspressoButton() {
	n.raspi.ActivateEspressoButton(true)
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)
	n.raspi.ActivateEspressoButton(false)
}

func (n NespressoMachine) pressLungoButton() {
	n.raspi.ActivateLungoButton(true)
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)
	n.raspi.ActivateLungoButton(false)
}

func (n NespressoMachine) MakeEspresso() {
	// switch on machine
	n.pressEspressoButton()
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)

	// make espresso
	n.pressEspressoButton()
}

func (n NespressoMachine) MakeLungo() {
	// switch on machine
	n.pressLungoButton()
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)

	// make lungo
	n.pressLungoButton()
}
