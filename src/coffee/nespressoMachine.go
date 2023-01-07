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

type nespressoMachine struct {
	raspi               raspberrypi
	buttonPressLengthMs int
}

func NewNespressoMachine(cfg NespressoMachineConfig, raspi raspberrypi) nespressoMachine {
	return nespressoMachine{raspi: raspi, buttonPressLengthMs: cfg.ButtonPressDurationMs}
}

func (n nespressoMachine) PressEspressoButton() {
	n.raspi.ActivateEspressoButton(true)
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)
	n.raspi.ActivateEspressoButton(false)
}

func (n nespressoMachine) PressLungoButton() {
	n.raspi.ActivateLungoButton(true)
	time.Sleep(time.Duration(n.buttonPressLengthMs) * time.Millisecond)
	n.raspi.ActivateLungoButton(false)
}
