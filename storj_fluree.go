package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fluree/storj-fluree/fluree"
	"github.com/fluree/storj-fluree/storj"
	"github.com/urfave/cli"
)

const dbConfigFile = "./config/db_property.json"
const storjConfigFile = "./config/storj_config.json"

var gbDEBUG = false

// Create command-line tool to read from CLI.
var app = cli.NewApp()

func setAppInfo() {
	app.Name = "Storj-Fluree Connector"
	app.Usage = "A Storj-Fluree connector. Upload and retrieve Fluree snapshot files on Storj."
	app.Version = "0.2.0"
}

// helper function to flag debug
func setDebug(debugVal bool) {
	gbDEBUG = debugVal
	fluree.DEBUG = debugVal
	storj.DEBUG = debugVal
}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}

	return false
}

// setCommands sets various command-line options for the app.
func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "snapshot",
			Aliases: []string{"sn"},
			Usage:   "Command to create snapshot for the Fluree database specified in the configuration",
			//\narguments-\n\t  fileName [optional] = provide full file name (with complete path) to configuration file for Fluree database, if this is not provided, use default location.
			// debug [optional] Optionally also include "debug" as an argument  example = ./storj_fluree snapshot debug ./config/db_property.json\n",
			Action: func(cliContext *cli.Context) error {
				var fullFileName = dbConfigFile

				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args().Get(i) == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args().Get(i)
						}
					}
				}

				flureeConfig, err := fluree.LoadFlureeConfiguration(fullFileName)

				if err != nil {
					log.Fatalf("fluree.LoadFlureeConfiguration: %s", err)
				}

				// Connect to Database and process data
				snapshot, err := fluree.CreateSnapshot(flureeConfig)

				if err != nil {
					log.Fatalf("fluree.CreateSnapshot: %s", err)
				} else {
					fmt.Printf("Created an snapshot for %s/%s - %s\n...Complete!\n", flureeConfig.Network, flureeConfig.DBID, snapshot)
				}

				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "Lists all available snapshots.",
			//\narguments-\n\t  fileName [optional] = provide full file name (with complete path) to configuration file for Fluree database, if this is not provided, use default location.
			// debug [optional] Optionally also include "debug" as an argument  example = ./storj_fluree list debug ./config/db_property.json\n",
			Action: func(cliContext *cli.Context) error {
				var fullFileName = dbConfigFile

				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args().Get(i) == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args().Get(i)
						}
					}
				}

				flureeConfig, err := fluree.LoadFlureeConfiguration(fullFileName)

				if err != nil {
					log.Fatalf("fluree.LoadFlureeConfiguration: %s", err)
				}

				// Connect to Database and process data
				snapshotList, err := fluree.ListSnapshots(flureeConfig)

				if err != nil {
					log.Fatalf("fluree.ListSnapshot: %s", err)
				} else {
					fmt.Printf("\nAvailable snapshots: \n")

					for _, snapshot := range snapshotList {
						fmt.Printf("%s\n", snapshot)
					}
				}

				return nil
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Command to read and parse JSON information about Storj network and upload sample JSON data",
			//\n arguments- 1. fileName [optional] = provåßide full file name (with complete path), storing Storj configuration information if this fileName is not given, then data is read from ./config/storj_config.json example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) error {

				// Default Storj configuration file name.
				var fullFileName = storjConfigFile

				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args().Get(i) == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args().Get(i)
						}
					}
				}

				// Sample database name and data to be uploaded
				dbName := "testdb"
				testData := []byte("{'testKey': 'testValue'}")

				if gbDEBUG {
					t := time.Now()
					time := t.Format("2006-01-02_15:04:05")
					var fileName = "test/uploaddata_" + time + ".json"

					err := ioutil.WriteFile(fileName, testData, 0644)
					if err != nil {
						fmt.Println("Error while writing to file ")
					}
				}

				fileName := "test.json"

				err := storj.ConnectStorjUploadData(fullFileName, []byte(testData), fileName, dbName)
				if err != nil {
					fmt.Println("Error while uploading data to the Storj bucket")
				}

				return nil
			},
		},
		{
			Name:    "store",
			Aliases: []string{"st"},
			Usage:   "Command to connect and place Fluree snapshot to given Storj Bucket in JSON format",
			//\n    arguments-\n      1. fileName [optional] = provide full file name (with complete path), storing Fluree properties in JSON format\n   if this fileName is not given, then data is read from ./config/db_property.json\n      2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n     if this fileName is not given, then data is read from ./config/storj_config.json\n    3. snapshotName [optional] = provide snapshot file name, if not given, then the latest snapshot is used. example = ./storj_mongodb s ./config/db_property.json ./config/storj_config.json\n",
			Action: func(cliContext *cli.Context) error {

				// Default configuration file names.
				var fullFileNameStorj = storjConfigFile
				var fullFileNameFlureeDB = dbConfigFile
				var snapshotName string

				// process arguments - Reading fileName from the command line.
				var foundFirstFileName = false
				var foundSecondFileName = false
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase debug is provided as argument.
						if cliContext.Args().Get(i) == "debug" {
							setDebug(true)
						} else {
							if !foundFirstFileName {
								fullFileNameFlureeDB = cliContext.Args().Get(i)
								foundFirstFileName = true
							} else if !foundSecondFileName {
								fullFileNameStorj = cliContext.Args().Get(i)
								foundSecondFileName = true
							} else {
								snapshotName = cliContext.Args().Get(i)
							}
						}
					}
				}

				flureeConfig, err := fluree.LoadFlureeConfiguration(fullFileNameFlureeDB)

				if err != nil {
					log.Fatalf("fluree.LoadFlureeConfig: %s", err)
				}

				snapshotList, err := fluree.ListSnapshots(flureeConfig)

				if err != nil {
					log.Fatalf("fluree.ListSnapshots: %s", err)
				}

				// If snapshotName provided, check that it is valid
				// else get the latest snapshot
				if snapshotName != "" {
					snapshotInList := contains(snapshotList, snapshotName)

					if snapshotInList == false {
						log.Fatalf("Snapshot provided not in snapshot list for this database. Provided: %s", snapshotName)
					}

				} else {
					lastSnapshot, err := fluree.GetLatestSnapshot(snapshotList)

					if err != nil {
						log.Fatalf("fluree.GetLatestSnapshot: %s", err)
					}

					snapshotName = lastSnapshot
					if gbDEBUG {
						fmt.Printf("The latest snapshot is %s\n", snapshotName)
					}
				}

				if gbDEBUG {
					fmt.Printf("Attempting to upload snapshot %s for %s/%s", snapshotName, flureeConfig.Network, flureeConfig.DBID)
				}

				data, err := fluree.ReadSnapshot(flureeConfig, snapshotName)

				if err != nil {
					log.Fatalf("fluree.ReadSnapshot: %s", err)
				}

				// Connecting to storj network for uploading data.
				dbname := flureeConfig.Network + "/" + flureeConfig.DBID

				err = storj.ConnectStorjUploadData(fullFileNameStorj, []byte(data), snapshotName, dbname)
				if err != nil {
					fmt.Println("Error while uploading data to bucket ", err)
				}

				return nil
			},
		},
	}
}

func main() {

	setAppInfo()
	setCommands()

	setDebug(false)

	err := app.Run(os.Args)

	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
