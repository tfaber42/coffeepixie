package main

// These are the libraries we are going to use
// Both "fmt" and "net" are part of the Go standard library
import (
	// "fmt" has methods for formatted I/O operations (like printing to the console)
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	// The "net/http" library has methods to implement HTTP clients and servers
	"net/http"

	"html/template"

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

	coffeeTimer.SetTriggerFunc(pixie.MakeEspresso)

	raspi.SetShowArmedStatusFunc(coffeeTimer.ShowArmedStatus)
	raspi.SetToggleArmedStatusFunc(coffeeTimer.ToggleArmedStatus)

	coffeeTimer.Arm()

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

	port := "3000"

	fs := http.FileServer(http.Dir("src/html/assets"))
	ph := pixieHandler{coffeeTimer: coffeeTimer, nespressoMachine: &pixie}

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.Handle("/", ph)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

var tpl = template.Must(template.ParseFiles("src/html/index.html"))

type pageData struct {
	EspressoChecked, LungoChecked, NoCoffeeChecked string
	TriggerTime                                    string
	Status                                         template.HTML
}

type pixieHandler struct {
	coffeeTimer      *coffee.CoffeeTimer
	nespressoMachine *coffee.NespressoMachine
}

func (ph pixieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//u, err := url.Parse(r.URL.String())
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*
			params := u.Query()
			triggerTime := params.Get("trigger-time")
			triggerType := params.Get("trigger-type")


		fmt.Println("triggerTime is: ", triggerTime)
		fmt.Println("triggerType is: ", triggerType)
	*/

	triggerTime := r.PostFormValue("trigger-time")
	triggerType := r.PostFormValue("trigger-type")

	var pd pageData

	// rudimentary string validation
	if len(strings.Split(triggerTime, ":")) == 2 {
		ph.coffeeTimer.SetTriggerTime(triggerTime)
		pd.TriggerTime = triggerTime
	} else {
		pd.TriggerTime = ph.coffeeTimer.GetTriggerTime()
	}

	switch triggerType {
	case "espresso":
		ph.coffeeTimer.SetTriggerFunc(ph.nespressoMachine.MakeEspresso)
		ph.coffeeTimer.Arm()
	case "lungo":
		ph.coffeeTimer.SetTriggerFunc(ph.nespressoMachine.MakeLungo)
		ph.coffeeTimer.Arm()
	case "none":
		ph.coffeeTimer.Disarm()
	default:
		if ph.coffeeTimer.IsArmed() {
			triggerType = "espresso"
		} else {

			triggerType = "none"
		}
	}
	ph.coffeeTimer.ShowArmedStatus()

	switch triggerType {
	case "espresso":
		pd.EspressoChecked = "checked"
		pd.Status = template.HTML(fmt.Sprintf("Pixie is making <b>ESPRESSO</b> at %s", pd.TriggerTime))
	case "lungo":
		pd.LungoChecked = "checked"
		pd.Status = template.HTML(fmt.Sprintf("Pixie is making <b>LUNGO</b> at %s", pd.TriggerTime))
	case "none":
		pd.NoCoffeeChecked = "checked"
		pd.Status = template.HTML("Pixie is NOT MAKING COFFEE")
	}

	tpl.Execute(w, pd)
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
