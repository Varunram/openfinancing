package main

import (
	"fmt"
	"log"
	"net/http"

	xlm "github.com/YaleOpenLab/openx/xlm"
)

// commands.go has a list of all the commands supported by the teller. This is intentionally
// tweaked and limited to ensure that this serves only data related to the specific project
// at hand

func WriteToHandler(w http.ResponseWriter, jsonString []byte) {
	w.Header().Add("Access-Control-Allow-Origin", "localhost")
	w.Header().Add("Access-Control-Allow-Methods", "GET")
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonString)
}

func ParseInput(input []string) {
	if len(input) == 0 {
		fmt.Println("List of commands: ping, receive, display, update")
		return
	}

	command := input[0]
	switch command {
	case "qq":
		// handler to quit and test the teller without hashing the state and committing two transactions
		// each time we start the teller
		log.Fatal("qq emergency exit")
	case "help":
		fmt.Println("List of commands: ping, receive, display, update")
		break
	case "ping":
		err := PingRpc()
		if err != nil {
			log.Println(err)
		}
	case "receive":
		if len(input) != 2 {
			fmt.Println("USAGE: receive xlm")
			return
		}
		err := xlm.GetXLM(RecpPublicKey)
		if err != nil {
			log.Println(err)
		}
	case "display":
		if len(input) < 2 {
			fmt.Println("USAGE: display <balance, info>")
			return
		}
		subcommand := input[1]
		switch subcommand {
		case "balance":
			if len(input) < 3 {
				fmt.Println("USAGE: display balance <xlm, asset>")
				return
			}

			subsubcommand := input[2]
			var balance string
			var err error
			ColorOutput("Displaying balance in "+subsubcommand+" for user: ", WhiteColor)

			switch subsubcommand {
			case "xlm":
				balance, err = xlm.GetNativeBalance(RecpPublicKey)
			default:
				balance, err = xlm.GetAssetBalance(RecpPublicKey, subcommand)
			}

			if err != nil {
				log.Println(err)
				return
			}
			ColorOutput(balance, MagentaColor)
		case "info":
			var err error
			LocalProject, err = GetLocalProjectDetails(LocalProjIndex)
			if err != nil {
				log.Println(err)
				break
			}
			fmt.Println("          PROJECT INDEX: ", LocalProject.Index)
			fmt.Println("          Panel Size: ", LocalProject.PanelSize)
			fmt.Println("          Total Value: ", LocalProject.TotalValue)
			fmt.Println("          Location: ", LocalProject.Location)
			fmt.Println("          Money Raised: ", LocalProject.MoneyRaised)
			fmt.Println("          Metadata: ", LocalProject.Metadata)
			fmt.Println("          Years: ", LocalProject.Years)
			fmt.Println("          Auction Type: ", LocalProject.AuctionType)
			fmt.Println("          Debt Asset Code: ", LocalProject.DebtAssetCode)
			fmt.Println("          Payback Asset Code: ", LocalProject.PaybackAssetCode)
			fmt.Println("          Balance Left: ", LocalProject.BalLeft)
			fmt.Println("          Date Initiated: ", LocalProject.DateInitiated)
			fmt.Println("          Date Last Paid: ", LocalProject.DateLastPaid)
		default:
			// handle defaults here
			log.Println("Invalid command or need more parameters")
		} // end of display
	case "update":
		if len(input) != 1 {
			fmt.Println("USAGE: update <state>")
			return
		}
		updateState()
	}
}
