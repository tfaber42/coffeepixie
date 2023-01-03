package main

// These are the libraries we are going to use
// Both "fmt" and "net" are part of the Go standard library
import (
	// "fmt" has methods for formatted I/O operations (like printing to the console)
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// The "net/http" library has methods to implement HTTP clients and servers
	"net/http"

	"github.com/tfaber42/coffeepixie/src/coffee"
	"gopkg.in/yaml.v2"
	"periph.io/x/host/v3"
)

type Config struct {
	NespressoMachine coffee.NespressoMachineConfig `yaml:"nespresso_machine"`
	Timer            coffee.CoffeeTimerConfig      `yaml:"timer"`
}

func main() {

	cfg := readConfig("config.yml")

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		fmt.Print("error init")
		log.Fatal(err)
	}

	pixie := coffee.NewNespressoMachine(cfg.NespressoMachine)
	defer pixie.Disconnect()

	coffeeTimer := coffee.NewCoffeeTimer(cfg.Timer)
	defer coffeeTimer.Disconnect()

	coffeeTimer.SetTriggerFunc(func() {
		pixie.PressEspressoButton()
		time.Sleep(300 * time.Millisecond)
		pixie.PressEspressoButton()
	})

	// Clean up on ctrl-c and turn lights out
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("SIGTERM received")
		pixie.Disconnect()
		coffeeTimer.Disconnect()
		os.Exit(0)
	}()

	// The "HandleFunc" method accepts a path and a function as arguments
	// (Yes, we can pass functions as arguments, and even trat them like variables in Go)
	// However, the handler function has to have the appropriate signature (as described by the "handler" function below)
	http.HandleFunc("/", handler)

	// After defining our server, we finally "listen and serve" on port 8080
	// The second argument is the handler, which we will come to later on, but for now it is left as nil,
	// and the handler defined above (in "HandleFunc") is used
	err := http.ListenAndServe(":8080", nil)

	fmt.Println(err)

}

func readConfig(fileName string) Config {
	var cfg Config
	cfgFile, err := os.Open(fileName)
	if err != nil {
		// set defaults and write file
		cfg.NespressoMachine = coffee.NespressoMachineConfigDefaults
		cfg.Timer = coffee.CoffeeTimerConfigDefaults

		cfgFile, err = os.Create("config.yml")
		if err != nil {
			log.Fatal(err)
		}
		defer cfgFile.Close()

		encoder := yaml.NewEncoder(cfgFile)
		err = encoder.Encode(&cfg)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		defer cfgFile.Close()

		decoder := yaml.NewDecoder(cfgFile)
		err = decoder.Decode(&cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
	return cfg
}

// "handler" is our handler function. It has to follow the function signature of a ResponseWriter and Request type
// as the arguments.
func handler(w http.ResponseWriter, r *http.Request) {
	// For this case, we will always pipe "Hello World" into the response writer
	fmt.Fprintf(w, "Hello World!")
}
