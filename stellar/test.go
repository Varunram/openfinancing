package main

// test.go runs the PoC stellar implementation calling various functions
import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	accounts "github.com/YaleOpenLab/smartPropertyMVP/stellar/accounts"
	assets "github.com/YaleOpenLab/smartPropertyMVP/stellar/assets"
	database "github.com/YaleOpenLab/smartPropertyMVP/stellar/database"
	rpc "github.com/YaleOpenLab/smartPropertyMVP/stellar/rpc"
	utils "github.com/YaleOpenLab/smartPropertyMVP/stellar/utils"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	// TOOD: define default values for each and then use them if not passed
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
	// InvAmount int    `short:"i" description:"Desired investment" required:"true"`
	InvAmount int `short:"i" description:"Desired investment"`
	// RecYears  int    `short:"r" description:"Number of years the recipient wants to repay in. Can be 3, 5 or 7 years." required:"true"`
	RecYears int    `short:"r" description:"Number of years the recipient wants to repay in. Can be 3, 5 or 7 years."`
	Port     string `short:"p" description:"The port on which the server runs on"`
}

func ValidateInputs() {
	if (opts.RecYears != 0) && !(opts.RecYears == 3 || opts.RecYears == 5 || opts.RecYears == 7) {
		// right now payoff periods are limited, I guess they don't need to be,
		// but in this case jsut are
		log.Fatal(fmt.Errorf("Number of years not supported"))
	}
}

func main() {
	var err error
	_, err = flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	// For the demo, we must have multiple things that are in line
	// 1. An interface to view the number of orders that are in the orderbook
	// 2. An interface to view all the assets owned by a particular investor
	// 3. An option to invest in a prticular option, guided by the CLI
	// Available orders: <display avilable orders here>
	// Choose which order to invest in: (don't have amount initially since we assume
	// that the investor is investing the whole of the amount required by the given order)
	// after investment, it should display the INVAsset's code, the hash of the asset
	// sending transaction and a confirmation that the investor has invested. Then it must display
	// something like <View invested Assets> which the person can click on to see
	// what he has invested in and there,  it must show the INVAsset and INVAmount

	// so we first need to display all invested assets
	// and maybe print something Stellar Housing Assets interface
	// clear db later, have this in for now

	fmt.Println("------------STELLAR HOUSE INVESTMENT CLI INTERFACE------------")
	ValidateInputs()

	// insert an investor with the relevant details
	// add dummy investor and recipient data for the demo
	// uname: john, password: password
	/*
		nInvestor, err := database.NewInvestor("john",
			"e9a75486736a550af4fea861e2378305c4a555a05094dee1dca2f68afea49cc3a50e8de6ea131ea521311f4d6fb054a146e8282f8e35ff2e6368c1a62e909716",
			"John", true)
		err = database.InsertInvestor(nInvestor)
		if err != nil {
			log.Fatal(err)
		}

		// need to add a dummy recipient here as well
		// uname: martin, password:password
		nRecipient, err := database.NewRecipientWithoutSeed("martin",
			"8a56bac869374c669443a1626ff0967af258123f83faf6b55e31dd541e6bbd90308a3385713294bf2e8861bc8cf8f8feda41f9c4db19d5811a6b5de85eac9870",
			"Martin")
		err = database.InsertRecipient(nRecipient)
		if err != nil {
			log.Fatal(err)
		}
	*/
	// need to ask for user role as well here, to know whether the user is an investor
	// or recipient so that we can show both sides
	// Open the database
	err = database.InsertDummyData()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("---ARE YOU AN INVESTOR (I) OR RECIPIENT (R)? ---")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	rbool := false
	if scanner.Text() == "I" || scanner.Text() == "i" {
		fmt.Println("WELCOME BACK INVESTOR")
	} else if scanner.Text() == "R" || scanner.Text() == "r" {
		fmt.Println("WELCOME BACK RECIPIENT")
		rbool = true
	}
	// ask for username and password combo here
	fmt.Println("---ENTER YOUR USERNAME---")
	scanner.Scan()
	if scanner.Err() != nil {
		fmt.Println("Couldn't read user input")
	}
	invLoginUserName := scanner.Text() // read user input regarding which option
	fmt.Println("---ENTER YOUR PASSWORD---")
	scanner.Scan()
	if scanner.Err() != nil {
		fmt.Println("Couldn't read user input")
	}
	invLoginPassword := utils.SHA3hash(scanner.Text())
	log.Println("INV PASSWORD IS: ", invLoginPassword, invLoginUserName)
	// check for ibool vs rbool here
	if rbool {
		// handle the recipient case here because its simpler
		dbRecipient, err := database.SearchForRecipientPassword(invLoginPassword)
		if err != nil {
			log.Fatal("had trouble retrieving the password")
		}
		if dbRecipient.LoginUserName != invLoginUserName { // should rework to check the password, this is just a temp hack
			log.Fatal("UserNames don't match: ", dbRecipient.LoginUserName, invLoginUserName, dbRecipient)
		}
		// at this point, we have verified the recipient
		// have a for loop here with various options
		for {
			fmt.Println("------------RECIPIENT INTERFACE------------")
			fmt.Println("----CHOOSE ONE OF THE FOLLOWING OPTIONS----")
			fmt.Println("  1. Display all Received Assets")
			fmt.Println("  2. Display my Profile")
			fmt.Println("  3. Payback towards an Asset")
			fmt.Println("  4. Exit")
			scanner.Scan()
			if scanner.Err() != nil {
				fmt.Println("Couldn't read user input")
			}
			menuInput, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Fatal(err)
			}
			switch menuInput {
			case 1:
				break
			case 2:
				database.PrettyPrintRecipient(dbRecipient)
				break
			case 3:
				// need to test this one out
				break
			case 4:
				// check whether he wants to go back to the display all screen again
				fmt.Println("DO YOU REALLY WANT TO EXIT? (PRESS Y TO CONFIRM)")
				scanner.Scan()
				if scanner.Text() == "Y" || scanner.Text() == "y" {
					fmt.Println("YOU HAVE DECIDED TO EXIT")
					log.Fatal("")
				}
				break
			}
		}
		database.PrettyPrintRecipient(dbRecipient)
		return
	}

	allOrders1, err := database.RetrieveAllOrdersWithoutDB()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ALL ORDERS", allOrders1)
	allInvestors, err := database.RetrieveAllInvestors()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ALLIVN", allInvestors)
	dbInvestor, err := database.SearchForInvestorPassword(invLoginPassword)
	if err != nil {
		log.Fatal("had trouble retrieving the password")
	}

	if dbInvestor.LoginUserName != invLoginUserName { // should rework to check the password, this is just a temp hack
		log.Println("INV USERNAME: ", invLoginUserName, "DB USERNAME: ", dbInvestor.LoginUserName)
		log.Fatal("Investor username doesn't match", dbInvestor.LoginUserName, invLoginUserName)
	}

	for {
		// Main investor loop
		// password is password
		// ask for the investor username and password
		/*
			err = database.DeleteOrder(1, db)
			if err != nil {
				log.Println("Error deleting entry 1 from database")
			}
			err = database.DeleteOrder(2, db)
			if err != nil {
				log.Println("Error deleting entry 2 from database")
			}
		*/
		// have a menu here that will ask you what you want to do
		fmt.Println("------------INVESTOR INTERFACE------------")
		fmt.Println("----CHOOSE ONE OF THE FOLLOWING OPTIONS----")
		fmt.Println("  1. Display all Open Assets")
		fmt.Println("  2. Display my Profile")
		fmt.Println("  3. Invest in an Asset")
		fmt.Println("  4. Exit")
		scanner.Scan()
		if scanner.Err() != nil {
			fmt.Println("Couldn't read user input")
		}
		menuInput, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Fatal(err)
		}
		switch menuInput {
		case 1:
			fmt.Println("------------LIST OF ALL AVAILABLE ASSETS------------")
			time.Sleep(1 * time.Second) // change this to 5 or something for the pause
			allOrders, err := database.RetrieveAllOrdersWithoutDB()
			if err != nil {
				log.Println("Error retrieving all orders from the database")
			}
			database.PrettyPrintOrders(allOrders)
			break
		case 2:
			database.PrettyPrintInvestor(dbInvestor)
			break
		case 3:
			fmt.Println("----WHICH ASSET DO YOU WANT TO INVEST IN? (ENTER NUMBER WITHOUT SPACES)----")
			scanner.Scan()
			if scanner.Err() != nil {
				fmt.Println("Couldn't read user input")
			}
			// they want to choose
			// also check whether received user input is an integer
			uInput, err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Fatal("user input is not a number")
			}
			// now the user has decided to invest in the asset with index uInput
			// we need to retrieve the order and ask for confirmation
			uOrder, err := database.RetrieveOrderRPC(uint32(uInput))
			if err != nil {
				log.Fatal("Order with specified index not found in the database")
			}
			database.PrettyPrintOrder(uOrder)
			fmt.Println(" DO YOU WANT TO CONFIRM THIS ORDER? (PRESS N IF YOU DON'T WANT TO)")
			scanner.Scan()
			if scanner.Text() == "N" || scanner.Text() == "n" {
				fmt.Println("YOU HAVE DECIDED TO CANCEL THIS ORDER")
				break
			}
			// now we need to setup the dummy assets, setup a receiver as well to whom we send
			// the debt tokens and stuff
			// we also need to store the resulting assets in the respective arrays and then
			// display final investor status
			// setup issuer account
			issuer := accounts.SetupAccount()
			// we assume a centralized investor account
			investor := accounts.SetupAccount()
			dbInvestor.Seed = investor.Seed
			dbInvestor.PublicKey = investor.PublicKey
			// create a recipient fro the school
			recipient := accounts.SetupAccount()
			// everyone should have coins to setup trustlines.
			// anyways, stellar has a fat testnet wallet, so no worry that this might
			// get depleted
			err = issuer.GetCoins() // get coins for issuer
			if err != nil {
				log.Fatal(err)
			}

			err = issuer.SetupAccount(recipient.PublicKey, "10")
			if err != nil {
				log.Println("Recipient Account not setup")
				log.Fatal(err)
			}

			err = issuer.SetupAccount(investor.PublicKey, "10")
			if err != nil {
				log.Println("Investor Account not setup")
				log.Fatal(err)
			}

			log.Println("The issuer's public key and private key are: ", issuer.PublicKey, " ", issuer.Seed)
			log.Println("The investor's public key and private key are: ", investor.PublicKey, " ", investor.Seed)
			log.Println("The recipient's public key and private key are: ", recipient.PublicKey, " ", recipient.Seed)

			// so now we have three entities setup, so we create the assets and invest in them
			cOrder, err := assets.SetupAsset(&issuer, &investor, &recipient, uOrder) // assume payback period is 5
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("YOUR ORDER HAS BEEN CONFIRMED")
			// need to close db before this
			database.PrettyPrintOrder(cOrder)
			// now at this point, we need to assign the specific DEBToken to the recipient
			// we need the recipient's name in the order itself
			// so we can find the recipient while searching through the database
			dbRecipient, err := database.SearchForRecipientName(cOrder.RecipientName) // our recipient of assets
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Updating investor to handle invested amounts and assets")
			dbInvestor.AmountInvested += float64(cOrder.TotalValue) // we assume a single investor here
			dbInvestor.InvestedAssets = append(dbInvestor.InvestedAssets, cOrder)
			// now update the database with the investor
			err = database.InsertInvestor(dbInvestor)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Updated investor database")
			dbRecipient.Seed = recipient.Seed
			dbRecipient.PublicKey = recipient.PublicKey
			dbRecipient.DebtAssets = append(dbRecipient.DebtAssets, cOrder.DEBAssetCode)
			dbRecipient.PaybackAssets = append(dbRecipient.PaybackAssets, cOrder.PBAssetCode)

			err = database.InsertRecipient(dbRecipient)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Updated recipient bucket")

			// also update investor's invested assets
			fmt.Println("PLEASE CHECK A BLOCKHAIN EXPLORER TO CONFIRM BALANCES TO CONFIRM: ")
			fmt.Println("https://testnet.steexp.com/account/" + investor.PublicKey + "#balances")
			break
		case 4:
			// check whether he wants to go back to the display all screen again
			fmt.Println("DO YOU REALLY WANT TO EXIT? (PRESS Y TO CONFIRM)")
			scanner.Scan()
			if scanner.Text() == "Y" || scanner.Text() == "y" {
				fmt.Println("YOU HAVE DECIDED TO EXIT")
				log.Fatal("")
			}
		} // end of switch
	}
	log.Fatal("")
	// we now have the order the user wants to confirm, pretty print this order
	log.Fatal("All good")
	// start the rpc server
	rpc.StartServer(opts.Port) // this must be towards the end
	/*
		// open and close the db only for testing
		// in later cases, use the RPC directly
		log.Printf("InvAmount: %d USD, RecYears: %d years, Verbose: %t", opts.InvAmount, opts.RecYears, opts.Verbose)

		// the problem with this is we generally accept donations in crypto and then
		// people have to trust this that we don't print stuff out of thin air
		// instead of using our own coin, we could use stronghold coin (stablecoin on Stellar)
		// Stellar also has an immediate DEX, but do we use it? ethical stuff while dealing with
		// funds remiain
		// before setting up the assets, we need to refer to the orderbook in order to
		// get the list of available offers and funding things. For this purpose, we could
		// build a hash table / a simple dictionary, but I think investors in general
		// would like more info, so a simple map should be enough.
		// And this needs to be stored in a database somewhere so that we don't lose this
		// data. Also need cryptographic proofs that this data is what it is, because
		// there is no concept of state in stellar. Is there a better way?
		a, err := assets.SetupAsset(db, &issuer, &investor, &recipient, uOrder)
		if err != nil {
			log.Fatal(err)
		}
		// In short, the recipient pays in DEBtokens and receives PBtokens in return

		// this checks for balance, would come into use later on to check if we sent
		// the right amomunt of money to the user
		// balances, err := recipient.GetAllBalances()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// now we need to simulate a situation where the recipient pays back a certain
		// portion of the funds
		// onboarding is omitted here, that's a bigger problem that we hopefully
		// can delegate to other parties like Neighborly
		// an alternate idea is that they can buy stellar and repay, if we choose to
		// take that route, we must use a coin on stellar as an anchor to receive this token.
		// in this way, we need to check native balance and then use the anchor
		// right now don't do that, but should do in future to solicit donations from
		// the community, who would be generally dealing in XLM (and not DEBtoken)

		// another idea is that you could speculate on DEBtoken by having a market
		// for it, that would reuqire to relax the flags a bit. Right now, we don't
		// use an authorization flag, but we should since we don't want alternate markets
		// to develop. If we do, don't set the flag
		paybackAmount := "210"
		err = recipient.Payback(db, a.Index, a.DEBAssetCode, issuer.PublicKey, paybackAmount)
		if err != nil {
			log.Println(err)
			log.Fatal(err)
		}
		// after this ,we must update the steuff on the server side and send a payback token
		// to let the user know that he has paid x amoutn of money.
		// this however, would be the money paid / money that has to be paid per month
		// in total, this should be payBackPeriod * 12

		paybackAmountF := utils.StringToFloat(paybackAmount)
		refundS := utils.FloatToString(paybackAmountF / accounts.PriceOracleInFloat())
		// weird conversion stuff, but have to since the amount should be in a string
		blockHeight, txHash, err := issuer.SendAsset(a.PBAssetCode, recipient.PublicKey, refundS)
		if err != nil {
			log.Println("Error while sending a payback token, notify help immediately")
			log.Fatal(err)
		}
		log.Println("Sent payback token to recipient", blockHeight, txHash)
		tOrder, err := database.RetrieveOrder(a.Index, db)
		if err != nil {
			log.Println("Error retrieving from db")
			log.Fatal(err)
		}
		log.Println("Test whether this was updated: ", tOrder)

		debtAssetBalance, err := recipient.GetAssetBalance(a.DEBAssetCode)
		if err != nil {
			log.Fatal(err)
		}

		pbAssetBalance, err := recipient.GetAssetBalance(a.PBAssetCode)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Debt balance: %s, Payback Balance: %s", debtAssetBalance, pbAssetBalance)

		/*
			confHeight, txHash, err := issuer.SendCoins(recipient.PublicKey, "3.34") // send some coins from the issuer to the recipient
			if err != nil {
				log.Fatal(err)
			}

			log.Println("Confirmation height is: ", confHeight, " and txHash is: ", txHash)

			asset := issuer.CreateAsset(assetName) // create the asset that we want

			trustLimit := "100" // trust only 100 barrels of oil from Petro
			err = recipient.TrustAsset(asset, trustLimit)
			if err != nil {
				log.Println("Trust limit is in the wrong format")
				log.Fatal(err)
			}

			err = issuer.SendAsset(assetName, recipient.PublicKey, "3.34")
			if err != nil {
				log.Fatal(err)
			}
	*/
}
