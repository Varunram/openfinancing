package main

import (
	"encoding/json"
	"fmt"
	"log"

	database "github.com/YaleOpenLab/openx/database"
	solar "github.com/YaleOpenLab/openx/platforms/opensolar"
	rpc "github.com/YaleOpenLab/openx/rpc"
	utils "github.com/YaleOpenLab/openx/utils"
	geo "github.com/martinlindhe/google-geolocate"
)

func GetLocation(mapskey string) string {
	// see https://developers.google.com/maps/documentation/geolocation/intro on how
	// to improve location accuracy
	client := geo.NewGoogleGeo(mapskey)
	res, _ := client.Geolocate()
	location := fmt.Sprintf("Lat%fLng%f", res.Lat, res.Lng) // some ranodm format, can be improved upon if necessary
	DeviceLocation = location
	return location
}

func PingRpc() error {
	// make a curl request out to lcoalhost and get the ping response
	data, err := rpc.GetRequest(ApiUrl + "/ping")
	if err != nil {
		return err
	}
	var x rpc.StatusResponse
	// now data is in byte, we need the other strucutre now
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	// the result would be the status of the platform
	ColorOutput("PLATFORM STATUS: "+utils.ItoS(x.Code), GreenColor)
	return nil
}

func GetProjectIndex(assetName string) (int, error) {
	data, err := rpc.GetRequest(ApiUrl + "/project/funded")
	if err != nil {
		log.Println(err)
		return -1, err
	}
	var x []solar.Project
	err = json.Unmarshal(data, &x)
	if err != nil {
		return -1, err
	}
	for _, elem := range x {
		if elem.DebtAssetCode == assetName {
			return elem.Index, nil
		}
	}
	return -1, fmt.Errorf("Not found")
}

func LoginToPlatForm(username string, pwhash string) error {
	data, err := rpc.GetRequest(ApiUrl + "/recipient/validate?" + "username=" + username + "&pwhash=" + pwhash)
	if err != nil {
		return err
	}
	var x database.Recipient
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	ColorOutput("AUTHENTICATED RECIPIENT", GreenColor)
	LocalRecipient = x
	return nil
}

func ProjectPayback(assetName string, amount string) error {
	// retrieve project index
	projIndexI, err := GetProjectIndex(assetName)
	if err != nil {
		return fmt.Errorf("Couldn't pay")
	}
	projIndex := utils.ItoS(projIndexI)
	data, err := rpc.GetRequest(ApiUrl + "/recipient/payback?" + "username=" + LocalRecipient.U.Username +
		"&pwhash=" + LocalRecipient.U.Pwhash + "&projIndex=" + projIndex + "&assetName=" + assetName + "&seedpwd=" +
		LocalSeedPwd + "&amount=" + amount)
	if err != nil {
		return err
	}
	var x rpc.StatusResponse
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	if x.Code == 200 {
		ColorOutput("PAID!", GreenColor)
		return nil
	}
	return fmt.Errorf("Errored out")
}

func SetDeviceId(username string, pwhash string, deviceId string) error {
	data, err := rpc.GetRequest(ApiUrl + "/recipient/deviceId?" + "username=" + username +
		"&pwhash=" + pwhash + "&deviceid=" + deviceId)
	if err != nil {
		return err
	}
	var x rpc.StatusResponse
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	if x.Code == 200 {
		ColorOutput("PAID!", GreenColor)
		return nil
	}
	return fmt.Errorf("Errored out, didn't receive 200")
}

func StoreStartTime() error {
	data, err := rpc.GetRequest(ApiUrl + "/recipient/startdevice?" + "username=" + LocalRecipient.U.Username +
		"&pwhash=" + LocalRecipient.U.Pwhash + "&start=" + utils.I64toS(utils.Unix()))
	if err != nil {
		return err
	}

	var x rpc.StatusResponse
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	if x.Code == 200 {
		ColorOutput("LOGGED START TIME SUCCESSFULLY!", GreenColor)
		return nil
	}
	return fmt.Errorf("Errored out, didn't receive 200")
}

func StoreLocation(mapskey string) error {
	location := GetLocation(mapskey)
	data, err := rpc.GetRequest(ApiUrl + "/recipient/storelocation?" + "username=" + LocalRecipient.U.Username +
		"&pwhash=" + LocalRecipient.U.Pwhash + "&location=" + location)
	if err != nil {
		return err
	}

	var x rpc.StatusResponse
	err = json.Unmarshal(data, &x)
	if err != nil {
		return err
	}
	if x.Code == 200 {
		ColorOutput("LOGGED LOCATION SUCCESSFULLY!", GreenColor)
		return nil
	}
	return fmt.Errorf("Errored out, didn't receive 200")
}
