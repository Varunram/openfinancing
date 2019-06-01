package main

import (
	"github.com/pkg/errors"
	"log"

	database "github.com/YaleOpenLab/openx/database"
	solar "github.com/YaleOpenLab/openx/platforms/opensolar"
	rpc "github.com/YaleOpenLab/openx/rpc"
	scan "github.com/YaleOpenLab/openx/scan"
	wallet "github.com/YaleOpenLab/openx/xlm/wallet"
)

// Login logs on to the platform
func Login(username string, pwhash string) (string, error) {
	var wString string
	data, err := rpc.GetRequest(ApiUrl + "/user/validate?" + "username=" + username + "&pwhash=" + pwhash)
	if err != nil {
		return wString, errors.Wrap(err, "validate request failed")
	}
	var x rpc.ValidateParams
	err = x.UnmarshalJSON(data)
	if err != nil {
		return wString, errors.Wrap(err, "could not unmarshal json")
	}
	switch x.Role {
	case "Investor":
		wString = "Investor"
		data, err = rpc.GetRequest(ApiUrl + "/investor/validate?" + "username=" + username + "&pwhash=" + pwhash)
		if err != nil {
			return wString, errors.Wrap(err, "could not call ivnestor validate function")
		}
		var inv database.Investor
		err = inv.UnmarshalJSON(data)
		if err != nil {
			return wString, errors.Wrap(err, "could not unmarshal json")
		}
		LocalInvestor = inv
		ColorOutput("ENTER YOUR SEEDPWD: ", CyanColor)
		LocalSeedPwd, err = scan.ScanRawPassword()
		if err != nil {
			return wString, errors.Wrap(err, "could not scan raw password")
		}
		LocalSeed, err = wallet.DecryptSeed(LocalInvestor.U.EncryptedSeed, LocalSeedPwd)
		if err != nil {
			return wString, errors.Wrap(err, "could not decrypt seed")
		}
	case "Recipient":
		wString = "Recipient"
		data, err = rpc.GetRequest(ApiUrl + "/recipient/validate?" + "username=" + username + "&pwhash=" + pwhash)
		if err != nil {
			return wString, errors.Wrap(err, "could not call recipient validate endpoint")
		}
		var recp database.Recipient
		err = recp.UnmarshalJSON(data)
		if err != nil {
			return wString, errors.Wrap(err, "could not unmarshal json")
		}
		LocalRecipient = recp
		ColorOutput("ENTER YOUR SEEDPWD: ", CyanColor)
		LocalSeedPwd, err = scan.ScanRawPassword()
		if err != nil {
			return wString, errors.Wrap(err, "could not scan raw password")
		}
		LocalSeed, err = wallet.DecryptSeed(LocalRecipient.U.EncryptedSeed, LocalSeedPwd)
		if err != nil {
			return wString, errors.Wrap(err, "could not decrypt seed")
		}
	case "Entity":
		log.Println("ENTITY?")
		data, err = rpc.GetRequest(ApiUrl + "/entity/validate?" + "username=" + username + "&pwhash=" + pwhash)
		if err != nil {
			return wString, errors.Wrap(err, "could not call entity validate endpoint")
		}
		var entity solar.Entity
		err = entity.UnmarshalJSON(data)
		if err != nil {
			return wString, errors.Wrap(err, "could not unmarshal json")
		}
		if entity.Contractor {
			LocalContractor = entity
			wString = "Contractor"
		} else if entity.Originator {
			LocalOriginator = entity
			wString = "Originator"
		} else {
			return wString, errors.New("Not a contractor")
		}
		ColorOutput("ENTER YOUR SEEDPWD: ", CyanColor)
		LocalSeedPwd, err = scan.ScanRawPassword()
		if err != nil {
			return wString, errors.Wrap(err, "could not scan raw password")
		}
		if entity.Contractor {
			LocalSeed, err = wallet.DecryptSeed(LocalContractor.U.EncryptedSeed, LocalSeedPwd)
			if err != nil {
				return wString, errors.Wrap(err, "could not decrypt seed")
			}
		} else if entity.Originator {
			LocalSeed, err = wallet.DecryptSeed(LocalOriginator.U.EncryptedSeed, LocalSeedPwd)
			if err != nil {
				return wString, errors.Wrap(err, "could not decrypt seed")
			}
		}
	}
	ColorOutput("AUTHENTICATED USER, YOUR ROLE IS: "+wString, GreenColor)
	return wString, nil
}
