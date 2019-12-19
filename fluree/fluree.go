package fluree

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// ConfigFluree depicts keys to search for within the db_property.json file.
type ConfigFluree struct {
	IP               string `json:"ip"`
	Network          string `json:"network"`
	DBID             string `json:"dbid"`
	StorageDirectory string `json:"storageDirectory"`
}

// LoadFlureeConfiguration reads and parses the JSON file that contain Fluree configuration information.
func LoadFlureeConfiguration(fullFileName string) (ConfigFluree, error) { // fullFileName for fetching Fluree credentials from  given JSON filename.

	var configFluree ConfigFluree

	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configFluree, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configFluree)

	// Display read information.
	fmt.Println("Read Fluree configuration from the", fullFileName, "file")
	fmt.Println("\nIP:", configFluree.IP)
	fmt.Println("Network:", configFluree.Network)
	fmt.Println("DBID:", configFluree.DBID)
	fmt.Println("Storage Directory:", configFluree.StorageDirectory)

	return configFluree, nil
}

// Creates a new Fluree db snaphot for the configured db
func CreateSnapshot(configFluree ConfigFluree) (string, error) {

	var url = configFluree.IP + "fdb/" + configFluree.Network + "/" + configFluree.DBID + "/snapshot"

	if DEBUG {
		fmt.Printf("Sending snapshot request to: %s\n", url)
	}

	resp, err := http.Post(url, "application/json", nil)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return string(body), err
}

// Returns a list of snapshots for the configured db
func ListSnapshots(configFluree ConfigFluree) ([]string, error) {
	snapshotDirectory := configFluree.StorageDirectory + "/" + configFluree.Network + "/" + configFluree.DBID + "/snapshot"
	snapshotList := []string{}

	files, err := ioutil.ReadDir(snapshotDirectory)
	if err != nil {
		return []string{}, err
	}

	for _, file := range files {
		snapshotList = append(snapshotList, file.Name())
	}

	return snapshotList, nil
}

func GetLatestSnapshot(snapshotList []string) (string, error) {
	latestSnapshot := 0

	for _, snapshot := range snapshotList {
		var snapshotName = strings.Split(snapshot, ".avro")[0]
		var snapshotNum, err = strconv.Atoi(snapshotName)

		if err != nil {
			snapshotNum = 0
		}

		if snapshotNum > latestSnapshot {
			latestSnapshot = snapshotNum
		}
	}

	latestSnapshotName := strconv.Itoa(latestSnapshot) + ".avro"

	return latestSnapshotName, nil
}

func ReadSnapshot(configFluree ConfigFluree, snapshotName string) ([]byte, error) {
	snapshotFilePath := configFluree.StorageDirectory + "/" + configFluree.Network + "/" + configFluree.DBID + "/snapshot/" + snapshotName

	return ioutil.ReadFile(snapshotFilePath)
}
