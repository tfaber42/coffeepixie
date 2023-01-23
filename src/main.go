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

	"github.com/gorilla/mux"
	"github.com/tfaber42/coffeepixie/src/coffee"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	RaspberryPi      coffee.RaspiConfig            `yaml:"raspberry_pi"`
	NespressoMachine coffee.NespressoMachineConfig `yaml:"nespresso_machine"`
	Timer            coffee.CoffeeTimerConfig      `yaml:"timer"`
}

func main() {

	log.SetFlags(log.Flags() | log.Lmicroseconds)
	log.SetOutput(&lumberjack.Logger{
		Filename:   "coffeepixie.log",
		MaxSize:    5, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	})
	log.Println("***** Starting up coffee pixie *****")

	cfg := readConfig("config.yml")

	raspi := coffee.NewRaspi(cfg.RaspberryPi)
	defer raspi.Disconnect()

	pixie := coffee.NewNespressoMachine(cfg.NespressoMachine, raspi)

	coffeeTimer := coffee.NewCoffeeTimer(cfg.Timer, raspi)

	coffeeTimer.SetTriggerFunc(func() {
		pixie.PressEspressoButton()
		time.Sleep(300 * time.Millisecond)
		pixie.PressEspressoButton()
	})

	raspi.SetShowArmedStatusFunc(coffeeTimer.ShowArmedStatus)
	raspi.SetToggleArmedStatusFunc(coffeeTimer.ToggleArmedStatus)

	coffeeTimer.ShowArmedStatus()

	// Clean up on ctrl-c and turn lights out
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("SIGTERM received")
		raspi.Disconnect()
		log.Println()
		log.Println()
		os.Exit(0)
	}()

	// The "HandleFunc" method accepts a path and a function as arguments
	// (Yes, we can pass functions as arguments, and even trat them like variables in Go)
	// However, the handler function has to have the appropriate signature (as described by the "handler" function below)
	//http.HandleFunc("/", handler)

	// After defining our server, we finally "listen and serve" on port 8080
	// The second argument is the handler, which we will come to later on, but for now it is left as nil,
	// and the handler defined above (in "HandleFunc") is used
	err := http.ListenAndServe(":8080", newHttpRouter())

	fmt.Println(err)

}

func readConfig(fileName string) Config {
	var cfg Config
	cfgFile, err := os.Open(fileName)
	if err != nil {
		// set defaults and write file

		// TODO: reactivate the input buttons when the hardware interference issue is resolved
		cfg.RaspberryPi = coffee.RaspiConfigNoInputButtons

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

func newHttpRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/hello", handler).Methods("GET")

	staticFileDirectory := http.Dir("./html/")
	// Declare the handler, that routes requests to their respective filename.
	// The fileserver is wrapped in the `stripPrefix` method, because we want to
	// remove the "/assets/" prefix when looking for files.
	// For example, if we type "/assets/index.html" in our browser, the file server
	// will look for only "index.html" inside the directory declared above.
	// If we did not strip the prefix, the file server would look for
	// "./assets/assets/index.html", and yield an error
	staticFileHandler := http.StripPrefix("/html/", http.FileServer(staticFileDirectory))
	// The "PathPrefix" method acts as a matcher, and matches all routes starting
	// with "/assets/", instead of the absolute route itself
	r.PathPrefix("/html/").Handler(staticFileHandler).Methods("GET")

	return r
}
