package accounts

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/stellar/go/build"
	clients "github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/protocols/horizon" // using this since hte client/horizon package has some deprecated fields
)

type Account struct {
	Seed      string
	PublicKey string
}

var DefaultTestNetClient = &clients.Client{
	URL:  "https://horizon-testnet.stellar.org",
	HTTP: http.DefaultClient,
}

// Setup Account is a handler to setup a new account (issuer / investor / school)
func SetupAccount() Account {
	a, err := New()
	if err != nil {
		log.Fatal(err)
	}
	return a
}

func New() (Account, error) {
	var a Account
	pair, err := keypair.Random()
	// so key value pairs over here are ed25519 key pairs instead of bitcoin style key pairs
	// they also seem to sue al lcaps, which I don't know why
	// friendbot creates the account for us, on mainnet, there is no friendbot, so
	// we need to fil an address and then create accounts from that
	if err != nil {
		return a, err
	}
	log.Println("MY SEED IS: ", pair.Seed())
	a.Seed = pair.Seed()
	a.PublicKey = pair.Address()
	return a, nil
}

func (issuer *Account) SetupAccount(recipientPubKey string, amount string) (error){
	passphrase := network.TestNetworkPassphrase
	tx, err := build.Transaction(
		build.SourceAccount{issuer.Seed},
		build.AutoSequence{DefaultTestNetClient},
		build.Network{passphrase},
		build.CreateAccount(
			build.Destination{recipientPubKey},
			build.NativeAmount{amount},
		),
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	txe, err := tx.Sign(issuer.Seed)
	if err != nil {
		fmt.Println(err)
		return err
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return err
	}

	resp, err := DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)
	log.Println("LEDGER: ", resp.Ledger, "Hash: ", resp.Hash)
	return nil
}

func (a *Account) GetCoins() error {
	// get some coins from the stellar robot for testing
	// gives only a constant amount of stellar, so no need to pass it a coin param
	resp, err := http.Get("https://friendbot.stellar.org/?addr=" + a.PublicKey)
	if err != nil || resp == nil {
		log.Println("ERRORED OUT while calling friendbot, no coins for us")
		return err
	}
	return nil
}

func (a *Account) GetAssetBalance(assetCode string) (string, error) {

	account, err := DefaultTestNetClient.LoadAccount(a.PublicKey)
	if err != nil {
		return "", nil
	}

	for _, balance := range account.Balances {
		if balance.Asset.Code == assetCode {
			return balance.Balance, nil
		}
	}

	return "", nil
}

func (a *Account) GetAllBalances() ([]horizon.Balance, error) {

	account, err := DefaultTestNetClient.LoadAccount(a.PublicKey)
	if err != nil {
		return nil, nil
	}

	return account.Balances, nil
}

func (a *Account) SendCoins(destination string, amount string) (int32, string, error) {

	if _, err := DefaultTestNetClient.LoadAccount(destination); err != nil {
		// if destination doesn't exist, do nothing
		// returning -11 since -1 maybe returned for unconfirmed tx or something like that
		return -11, "", err
	}

	passphrase := network.TestNetworkPassphrase

	tx, err := build.Transaction(
		build.Network{passphrase},
		build.SourceAccount{a.Seed},
		build.AutoSequence{DefaultTestNetClient},
		build.Payment(
			build.Destination{destination},
			build.NativeAmount{amount},
		),
	)

	if err != nil {
		return -11, "", err
	}

	// Sign the transaction to prove you are actually the person sending it.
	txe, err := tx.Sign(a.Seed)
	if err != nil {
		return -11, "", err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		return -11, "", err
	}
	// And finally, send it off to Stellar!
	resp, err := DefaultTestNetClient.SubmitTransaction(txeB64)
	if err != nil {
		return -11, "", err
	}

	fmt.Println("Successful Transaction:")
	fmt.Println("Ledger:", resp.Ledger)
	fmt.Println("Hash:", resp.Hash)
	return resp.Ledger, resp.Hash, nil
}

func (a *Account) CreateAsset(assetName string) build.Asset {
	// need to set a couple flags here
	return build.CreditAsset(assetName, a.PublicKey)
}

func (a *Account) TrustAsset(asset build.Asset, limit string) (string, error) {
	// TRUST is FROM recipient TO issuer
	trustTx, err := build.Transaction(
		build.SourceAccount{a.PublicKey},
		build.AutoSequence{SequenceProvider: DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(asset.Code, asset.Issuer, build.Limit(limit)),
	)

	if err != nil {
		return "", err
	}

	trustTxe, err := trustTx.Sign(a.Seed)
	if err != nil {
		return "", err
	}

	trustTxeB64, err := trustTxe.Base64()
	if err != nil {
		return "", err
	}

	tx, err := DefaultTestNetClient.SubmitTransaction(trustTxeB64)
	if err != nil {
		return "", err
	}

	log.Println("Trusted asset tx: ", tx.Hash)
	return tx.Hash, nil
}

func (a *Account) SendAsset(assetName string, destination string, amount string) (int32, string, error) {
	// this transaction is FROM issuer TO recipient
	paymentTx, err := build.Transaction(
		build.SourceAccount{a.PublicKey},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: destination},
			build.CreditAmount{assetName, a.PublicKey, amount},
			// CreditAmount identifies the asset by asset Code and issuer pubkey
		),
	)

	if err != nil {
		return -11, "", err
	}

	paymentTxe, err := paymentTx.Sign(a.Seed)
	if err != nil {
		return -11, "", err
	}

	paymentTxeB64, err := paymentTxe.Base64()
	if err != nil {
		return -11, "", err
	}

	tx, err := DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
	if err != nil {
		return -11, "", err
	}

	return tx.Ledger, tx.Hash, nil
}

func (a *Account) SendAssetToIssuer(assetName string, issuerPubkey string, amount string) (int32, string, error) {
	// SendAssetToIssuer is FROM recipient / investor to issuer
	paymentTx, err := build.Transaction(
		build.SourceAccount{a.PublicKey},
		build.TestNetwork,
		build.AutoSequence{SequenceProvider: DefaultTestNetClient},
		build.Payment(
			build.Destination{AddressOrSeed: issuerPubkey},
			build.CreditAmount{assetName, issuerPubkey, amount},
		),
	)

	if err != nil {
		return -11, "", err
	}

	paymentTxe, err := paymentTx.Sign(a.Seed)
	if err != nil {
		return -11, "", err
	}

	paymentTxeB64, err := paymentTxe.Base64()
	if err != nil {
		return -11, "", err
	}

	tx, err := DefaultTestNetClient.SubmitTransaction(paymentTxeB64)
	if err != nil {
		return -11, "", err
	}

	return tx.Ledger, tx.Hash, nil
}

func PriceOracle(assetName string) (string, error) {
	// this is where we must call the oracle to check power tariffs and similar
	// right now, we can hardcode this, but ideally this must call the website
	// that  is there in the ETH contract, should be easy to do, but hardcode
	// for now
	return "200", nil
}

func (a*Account) Payback(assetName string, issuerPubkey string, amount string) (error) {
	// this will be called by the recipient
	oldBalance, err := a.GetAssetBalance(assetName)
	if err != nil {
		log.Fatal(err)
	}

	PBAmount, err := PriceOracle(assetName)
	if err != nil {
		log.Println("Unable to fetch oracle price, exiting")
		return err
	}
	// the oracke needs to know the assetName so that it can find the other details
	// about this asset from the db. This should run on the server side and must
	// be split when we do run client side stuff.
	// hardcode for now, need to add the oracle here so that we
	// can do this dynamically
	confHeight, txHash, err := a.SendAssetToIssuer(assetName, issuerPubkey, amount)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Paid debt amount: ", amount, " back to issuer, tx hash: ", txHash, " ", confHeight)
	log.Println("Checking balance to see if our account was debited")
	newBalance, err := a.GetAssetBalance(assetName)
	if err != nil {
		log.Fatal(err)
	}

	newBalanceFloat, err := strconv.ParseFloat(newBalance, 32) // 32 bit floats
	if err != nil {
		log.Println(err)
		return err
	}
	oldBalanceFloat, err := strconv.ParseFloat(oldBalance, 32)
	if err != nil {
		log.Println(err)
		return err
	}
	amountFloat, err := strconv.ParseFloat(PBAmount, 32)
	if err != nil {
		log.Println(err)
		return err
	}

	paidAmount := oldBalanceFloat - newBalanceFloat
	log.Println("Old Balance: ", oldBalanceFloat, "New Balance: ", newBalanceFloat, "Paid: ", paidAmount, "Req Amount: ", amountFloat)

	if paidAmount < amountFloat {
		log.Println("Amount paid is less than amount required, please pay more")
		return fmt.Errorf("Amount paid is less than amount required, please pay more")
	} else if paidAmount > amountFloat {
		log.Println("You've chosen to pay more than what is required for this month. Adjusting payback period accordingly")
		return nil
	} else {
		log.Println("You've paid exactly what is required for this month. Payback period remains as usual")
		return nil
	}

	return nil
}
