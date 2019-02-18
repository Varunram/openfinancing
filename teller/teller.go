package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"log"
	"os"
	"os/signal"
	"strings"

	consts "github.com/YaleOpenLab/openx/consts"
	database "github.com/YaleOpenLab/openx/database"
	solar "github.com/YaleOpenLab/openx/platforms/opensolar"
)

// package teller contains the remote client code that would be run on the client's
// side and communicate information with us and with atonomi and other partners.
// that it belongs, the contract, recipient, and eg. person who installed it.
// Consider doing this with IoT partners, eg. Atonomi.

// Teller authenticates with the platform using a remote API and then retrieves
// credentials once authenticated. Both the teller and the project recipient on the
// platform are the same entity, just that the teller is associated with the hw device.
// hw device needs an id and stuff, hopefully Atonomi can give us that.
// Teller tracks whenever the device starts and goes off, so we know when exactly the device was
// switched off. This is enough as proof that the device was running in between. This also
// avoids needing to poll the blockchain often and saves on the (minimal, still) tx fee.

// Since we can't compile this directly on the raspberry pi, we need to cross compile he
// go executable and transfer it over to the raspberry pi
// the following should do the trick for us
// env GOOS=linux GOARCH=arm GOARM=5 go build
var (
	LocalRecipient database.Recipient
	LocalProject   solar.Project
	LocalProjIndex string
	LocalSeedPwd   string

	RecpSeed      string
	RecpPublicKey string

	PlatformPublicKey string
	PlatformEmail     string

	ApiUrl string

	DeviceId       string
	DeviceLocation string
	DeviceInfo     string

	StartHash string
	NowHash   string

	HashChainHeader string
)

var cleanupDone chan struct{}

func main() {
	var err error
	err = StartTeller() // login to the platform, set device id, etc
	if err != nil {
		log.Println("Failed to start teller", err)
		log.Fatal(err)
	}
	ColorOutput("TELLER PUBKEY: "+RecpPublicKey, GreenColor)
	ColorOutput("DEVICE ID: "+DeviceId, GreenColor)
	// channels for preventing immediate sigint. Need this so that the action of any party which attempts
	// to close the teller would still be reported to the platform and emailed to the recipient
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		fmt.Println("\nSigint received in end handler. not quitting!")
		close(cleanupDone)
	}()

	StartHash, err = BlockStamp()
	if err != nil {
		log.Fatal(err)
	}

	// run goroutines in the background to routinely check for payback and update state
	go checkPayback()
	go updateState()
	go startServer()
	go storeDataLocal()

	promptColor := color.New(color.FgHiYellow).SprintFunc()
	whiteColor := color.New(color.FgHiWhite).SprintFunc()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      promptColor("teller") + whiteColor("# "),
		HistoryFile: consts.TellerHomeDir + "/history.txt",
	})
	defer rl.Close()

	for {
		// setup reader with max 4K input chars
		msg, err := rl.Readline()
		if err != nil {
			var err error
			err = endHandler() // error, user wants to quit
			for err != nil {
				log.Println(err)
				err = endHandler()
				<-cleanupDone // to prevent user from quitting when endhandler is running
			}
			break
		}
		msg = strings.TrimSpace(msg)
		if len(msg) == 0 {
			continue
		}
		rl.SaveHistory(msg)

		cmdslice := strings.Fields(msg)
		ColorOutput("entered command: "+msg, YellowColor)

		ParseInput(cmdslice)
	}
}
