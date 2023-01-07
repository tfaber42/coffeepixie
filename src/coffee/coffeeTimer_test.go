package coffee

import (
	"testing"
	"time"
)

func TestArmedTrigger(t *testing.T) {

	dummyRaspi := NewRaspi(NoRaspiInUseConfig)
	ct := NewCoffeeTimer(CoffeeTimerConfigDefaults, dummyRaspi)

	// set up monitoring channel for trigger function
	triggerFuncChannel := make(chan bool)
	ct.SetTriggerFunc(func() { triggerFuncChannel <- true })

	// set up trigger to be as soon as possible (currently finest granularity is minute)
	testTriggerTime := time.Now().Add(1 * time.Second).Format("3:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// arm the trigger
	ct.arm()

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
	testTriggerTime := time.Now().Add(1 * time.Second).Format("3:04:05")
	ct.SetTriggerTime(testTriggerTime)

	// arm the trigger
	ct.disarm()

	// set up monitoring timer that sends "false" to the channel after a timeout
	timeout := time.Now().Add(2 * time.Second)
	ct.cancellableTimer = time.AfterFunc(time.Until(timeout), func() { triggerFuncChannel <- false })

	// check which one gets triggered first: the coffeeTime triggerFunc, or the timeout
	coffeeTimerHasTriggered := <-triggerFuncChannel

	if coffeeTimerHasTriggered {
		t.Fatal("coffeeTimer has not triggered though it was disarmed")
	}

}
