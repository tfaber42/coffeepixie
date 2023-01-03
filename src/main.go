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
	"periph.io/x/host/v3"
)

func main() {

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		fmt.Print("error init")
		log.Fatal(err)
	}

	pixie := coffee.NewNespressoMachine(27, 22)
	defer pixie.Disconnect()

	coffeeTimer := coffee.NewCoffeeTimer(17, 4, 24, 23)
	defer coffeeTimer.Disconnect()

	//coffeeTimer.SetTriggerTime(22, 36)

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
	err := http.ListenAndServe(":8081", nil)

	fmt.Println(err)

}

// "handler" is our handler function. It has to follow the function signature of a ResponseWriter and Request type
// as the arguments.
func handler(w http.ResponseWriter, r *http.Request) {
	// For this case, we will always pipe "Hello World" into the response writer
	fmt.Fprintf(w, "Hello World!")
}
