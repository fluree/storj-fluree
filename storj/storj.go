// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	// "bytes"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"storj.io/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

type ConfigStorj struct {
	ApiKey               string `json:"apikey"`
	SatelliteAddress     string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) (ConfigStorj, error) { // fullFileName for fetching storj V3 credentials from  given JSON filename.

	var configStorj ConfigStorj

	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)

	// Display read information.
	if DEBUG {
		fmt.Println("\nRead Storj configuration from ", fullFileName)
		fmt.Println("\nAPI Key\t\t: ", configStorj.ApiKey)
		fmt.Println("Satellite	: ", configStorj.SatelliteAddress)
		fmt.Println("Bucket		: ", configStorj.Bucket)
		fmt.Println("Upload Path\t: ", configStorj.UploadPath)
		fmt.Println("")
	}

	return configStorj, nil
}

// // ConnectStorjUploadData reads Storj configuration from given file,
// // connects to the desired Storj network, and
// // uploads given object to the desired bucket.

func ConnectStorjUploadData(fullFileName string, dataToUpload []byte, snapshotName string, databaseName string) error { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// dataToUpload contains data that will be uploaded to storj V3 network.
	// databaseName for adding dataBase name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	configStorj, err := LoadStorjConfiguration(fullFileName)

	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in storj.LoadStorjConfiguration ", err)
		}
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}

	// 1. First, we need to create a Config
	cfg := new(uplink.Config)
	// configure the User Agent
	cfg.UserAgent = "fluree"
	// is this in ms?
	cfg.DialTimeout = 1000

	// 2. Then, we request access with a passphrase
	ctx := context.Background()
	acs, err := uplink.RequestAccessWithPassphrase(ctx, configStorj.SatelliteAddress,
		configStorj.ApiKey, configStorj.EncryptionPassphrase)
	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in uplink.RequestAccessWithPassphrase: ", err)
		}
		return fmt.Errorf("uplink.RequestAccessWithPassphrase: %s", err)
	}

	if DEBUG {
		fmt.Println("\nSuccessfully requested access with passphrase")
	}

	// 3. Next, we open a project
	prj, err := uplink.OpenProject(ctx, acs)
	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in uplink.OpenProject: ", err)
		}
		return fmt.Errorf("uplink.OpenProject: %s", err)
	}

	if DEBUG {
		fmt.Println("\nSuccessfully opened project")
		fmt.Println("\nCreating file name in the bucket, as per current time...")
	}

	var filename = configStorj.UploadPath + databaseName + "_" + snapshotName
	configStorj.UploadPath = configStorj.UploadPath + filename

	opts := new(uplink.UploadOptions)

	// Expire in 30 seconds
	var t = time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var date = time.Date(year, month, day, hour, min, sec+30, 0, t.Location())
	opts.Expires = date

	upload, err := prj.UploadObject(ctx, configStorj.Bucket, filename, opts)

	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in project.UploadObject: ", err)
		}
		return fmt.Errorf("project.UploadObject: %s", err)
	}

	// Upload.Write returns the number of bytes written
	n, err := upload.Write(dataToUpload)
	fmt.Println("\n", n, " bytes written to bucket:", configStorj.Bucket, " File: ", filename)

	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in upload.Write: ", err)
		}
		return fmt.Errorf("upload.Write: %s", err)
	}

	err = upload.Commit()
	if err != nil {
		if DEBUG {
			fmt.Println("\nERROR in upload.Commit: ", err)
		}
		return fmt.Errorf("upload.Commit: %s", err)
	}

	if DEBUG {
		// Test uploaded data by downloading it.
		// Initiate a download of the same object again.
		// var downopts := new(uplink.DownloadOptions);
		var downopts = new(uplink.DownloadOptions)
		downopts.Length = -1

		download, err := prj.DownloadObject(ctx, configStorj.Bucket, filename, downopts)
		if err != nil {
			fmt.Println("Could not open download at ", filename, ": ", err)
			return fmt.Errorf("Could not open download at %q: %v", configStorj.UploadPath, err)
		}
		defer download.Close()

		fmt.Println("Downloading Object from bucket : Initiated....")
		// Read everything from the stream.
		receivedContents, err := ioutil.ReadAll(download)

		if err != nil {
			fmt.Println("Error: could not read object", err)
			return fmt.Errorf("could not read object: %v", err)
		}
		var fileNameDownload = "test/downloadeddata_" + ".avro"
		err = ioutil.WriteFile(fileNameDownload, receivedContents, 0644)

		if !bytes.Equal(dataToUpload, receivedContents) {
			fmt.Println("Error: uploaded data != downloaded data")
			return fmt.Errorf("error: uploaded data != downloaded data")
		}
		fmt.Println("Downloading Object from bucket : Complete!")
		fmt.Println("")
	}

	return nil
}
