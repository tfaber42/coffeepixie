package coffee

import (
	"testing"
	"time"
)

func TestTriggerIsRearmedFollowingSettingTriggerTime(t *testing.T) {

	dummyRaspi := NewRaspi(NoRaspiInUseConfig)
	ct := NewCoffeeTimer(CoffeeTimerConfigDefaults, dummyRaspi)

	// set up monitoring channel for trigger function
	triggerFuncChannel := make(chan bool)
	ct.SetTriggerFunc(func() { triggerFuncChannel <- true })

	// arm the trigger
	ct.Arm()

	// set up trigger to be as soon as possible (currently finest granularity is minute)
	testTriggerTime := time.Now().Add(1 * time.Second).Format("15:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// if trigger is armed already, setting the trigger time should automatically re-arm it

	// set up monitoring timer that sends "false" to the channel after a timeout
	timeout := time.Now().Add(2 * time.Second)
	ct.cancellableTimer = time.AfterFunc(time.Until(timeout), func() { triggerFuncChannel <- false })

	// check which one gets triggered first: the coffeeTime triggerFunc, or the timeout
	coffeeTimerHasTriggered := <-triggerFuncChannel

	if !coffeeTimerHasTriggered {
		t.Fatal("coffeeTimer has not triggered in time")
	}

}

func TestTriggerIsRearmedFollowingSettingTriggerFunc(t *testing.T) {

	dummyRaspi := NewRaspi(NoRaspiInUseConfig)
	ct := NewCoffeeTimer(CoffeeTimerConfigDefaults, dummyRaspi)

	// set up trigger to be as soon as possible (currently finest granularity is minute)
	testTriggerTime := time.Now().Add(1 * time.Second).Format("15:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// arm the trigger
	ct.Arm()

	// set up monitoring channel for trigger function
	triggerFuncChannel := make(chan bool)
	ct.SetTriggerFunc(func() { triggerFuncChannel <- true })

	// if trigger is armed already, setting the trigger func should automatically re-arm it

	// set up monitoring timer that sends "false" to the channel after a timeout
	timeout := time.Now().Add(2 * time.Second)
	ct.cancellableTimer = time.AfterFunc(time.Until(timeout), func() { triggerFuncChannel <- false })

	// check which one gets triggered first: the coffeeTime triggerFunc, or the timeout
	coffeeTimerHasTriggered := <-triggerFuncChannel

	if !coffeeTimerHasTriggered {
		t.Fatal("coffeeTimer has not triggered in time")
	}

}

func TestArmedTrigger(t *testing.T) {

	dummyRaspi := NewRaspi(NoRaspiInUseConfig)
	ct := NewCoffeeTimer(CoffeeTimerConfigDefaults, dummyRaspi)

	// set up monitoring channel for trigger function
	triggerFuncChannel := make(chan bool)
	ct.SetTriggerFunc(func() { triggerFuncChannel <- true })

	// set up trigger to be as soon as possible (currently finest granularity is minute)
	testTriggerTime := time.Now().Add(1 * time.Second).Format("15:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// arm the trigger
	ct.Arm()

	// set up monitoring timer that sends "false" to the channel after a timeout
	timeout := time.Now().Add(2 * time.Second)
	ct.cancellableTimer = time.AfterFunc(time.Until(timeout), func() { triggerFuncChannel <- false })

	// check which one gets triggered first: the coffeeTime triggerFunc, or the timeout
	coffeeTimerHasTriggered := <-triggerFuncChannel

	if !coffeeTimerHasTriggered {
		t.Fatal("coffeeTimer has not triggered in time")
	}

}

func TestDisarmedTrigger(t *testing.T) {

	dummyRaspi := NewRaspi(NoRaspiInUseConfig)
	ct := NewCoffeeTimer(CoffeeTimerConfigDefaults, dummyRaspi)

	// set up monitoring channel for trigger function
	triggerFuncChannel := make(chan bool)
	ct.SetTriggerFunc(func() { triggerFuncChannel <- true })

	// set up trigger to be as soon as possible (currently finest granularity is minute)
	testTriggerTime := time.Now().Add(1 * time.Second).Format("15:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// arm the trigger
	ct.Disarm()

	// set up monitoring timer that sends "false" to the channel after a timeout
	timeout := time.Now().Add(2 * time.Second)
	ct.cancellableTimer = time.AfterFunc(time.Until(timeout), func() { triggerFuncChannel <- false })

	// check which one gets triggered first: the coffeeTime triggerFunc, or the timeout
	coffeeTimerHasTriggered := <-triggerFuncChannel

	if coffeeTimerHasTriggered {
		t.Fatal("coffeeTimer has triggered though it was disarmed")
	}

}
