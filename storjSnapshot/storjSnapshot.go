// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storjSnapshot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"storj.io/storj/lib/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigStorj depicts keys to search for within the storj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
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
	fmt.Println("Read Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)

	return configStorj, nil
}

// ConnectStorjUploadData reads Storj configuration from given file,
// connects to the desired Storj network, and
// uploads given object to the desired bucket.

func ConnectStorjUploadData(fullFileName string, dataToUpload []byte, snapshotName string, databaseName string) error { // fullFileName for fetching storj V3 credentials from  given JSON filename
	// dataToUpload contains data that will be uploaded to storj V3 network.
	// databaseName for adding dataBase name in storj V3 filename.
	// Read Storj bucket's configuration from an external file.
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}

	fmt.Println("\nCreating New Uplink...")

	var cfg uplink.Config
	// configure the partner id
	cfg.Volatile.PartnerID = "a1ba07a4-e095-4a43-914c-1d56c9ff5afd"

	ctx := context.Background()

	uplinkstorj, err := uplink.NewUplink(ctx, &cfg)
	if err != nil {
		return fmt.Errorf("Could not create new Uplink object: %s", err)
	}
	defer uplinkstorj.Close()

	fmt.Println("Parsing the API key...")
	key, err := uplink.ParseAPIKey(configStorj.APIKey)
	if err != nil {
		return fmt.Errorf("Could not parse API key: %s", err)
	}

	if DEBUG {
		fmt.Println("API key \t   :", key)
		fmt.Println("Serialized API key :", key.Serialize())
	}

	fmt.Println("Opening Project...")
	proj, err := uplinkstorj.OpenProject(ctx, configStorj.Satellite, key)

	if err != nil {
		return fmt.Errorf("Could not open project: %s", err)
	}
	defer proj.Close()

	// Creating an encryption key from encryption passphrase.
	if DEBUG {
		fmt.Println("\nGetting encryption key from pass phrase...")
	}

	encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
	if err != nil {
		return fmt.Errorf("Could not create encryption key: %s", err)
	}

	// Creating an encryption context.
	access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)
	fmt.Println("Encryption access \t:", *access)

	// Serializing the parsed access, so as to compare with the original key.
	serializedAccess, err := access.Serialize()
	if err != nil {
		fmt.Println("Error Serialized key : ", err)
	}

	if DEBUG {
		fmt.Println("Serialized access key\t:", serializedAccess)
	}
	fmt.Println("Opening Bucket: ", configStorj.Bucket)

	// Open up the desired Bucket within the Project.
	bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, access)
	//
	if err != nil {
		return fmt.Errorf("Could not open bucket %q: %s", configStorj.Bucket, err)
	}
	defer bucket.Close()

	//fmt.Println("Getting data into a buffer...")
	buf := bytes.NewBuffer(dataToUpload)

	//fmt.Println("Creating file name in the bucket, as per current time...")
	var filename = databaseName + "_" + snapshotName
	configStorj.UploadPath = configStorj.UploadPath + filename

	fmt.Println("File path: ", configStorj.UploadPath)
	fmt.Println("Uploading of the object to the Storj bucket: Initiated...")

	// Uploading JSON to Storj.
	err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, nil)
	if err != nil {
		fmt.Println("Uploading of data failed :\n ", err)
		fmt.Println("\nRetrying to Uploading data .....")
		err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, nil)
		if err != nil {
			return fmt.Errorf("Could not upload: %s", err)
		}
	}

	fmt.Println("Uploading of the object to the Storj bucket: Completed!")

	if DEBUG {
		// Test uploaded data by downloading it.
		// serializedAccess, err := access.Serialize().
		// Initiate a download of the same object again.
		readBack, err := bucket.OpenObject(ctx, configStorj.UploadPath)
		if err != nil {
			return fmt.Errorf("could not open object at %q: %v", configStorj.UploadPath, err)
		}
		defer readBack.Close()

		fmt.Println("Downloading range")
		// We want the whole thing, so range from 0 to -1.
		strm, err := readBack.DownloadRange(ctx, 0, -1)
		if err != nil {
			return fmt.Errorf("could not initiate download: %v", err)
		}
		defer strm.Close()
		fmt.Println("Downloading Object from bucket : Initiated....")
		// Read everything from the stream.
		receivedContents, err := ioutil.ReadAll(strm)
		if err != nil {
			return fmt.Errorf("could not read object: %v", err)
		}
		var fileNameDownload = "downloadeddata_" + ".avro"
		err = ioutil.WriteFile(fileNameDownload, receivedContents, 0644)

		if !bytes.Equal(dataToUpload, receivedContents) {
			return fmt.Errorf("error: uploaded data != downloaded data")
		}
		fmt.Println("Downloading Object from bucket : Complete!")
	}

	return nil
}
