# Storj-Fluree

This project provides a connection between FlureeDB and the decentrailed cloud storage network, Storj. The functions in this tool allow a FlureeDB snapshot to be streamed to the Storj network. The `.avro` file containing the Fluree snapshot can later be retrieved from Storj and used to create a new database from the snapshot file. Using the first version of the connector framework, this project is provided as-is and without warranty.

This is designed for Fluree 0.11.0 and above and will not work with earlier verions.

## Initial Set-up

Make sure your `PATH` includes the `$GOPATH/bin` directory, so that your commands can be easily used [Refer: Install the Go Tools:](https://golang.org/doc/install)
```
export PATH=$PATH:$GOPATH/bin
```

Install [github.com/urfave/cli](https://github.com/urfave/cli), by running: 
```
$ go get github.com/urfave/cli
```
Install [storj-uplink](https://godoc.org/storj.io/storj/lib/uplink) go package, by running:
```
$ go get storj.io/storj/lib/uplink
```

## Configure Packages
```
$ chmod 555 configure.sh
$ ./configure.sh
```

## Build ONCE
```
$ go build storj_fluree.go
```


## Set-up Files
* Create a `db_property.json` file, with following contents about a FlureeDB instance:
```json
    {
        "ip": "http://localhost:8090/",
        "network": "fluree",
        "dbid": "test",
        "storageDirectory": "/FULL/PATH/TO/STORAGE/FOLDER/data/ledger"
    }
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
```json
    { 
        "apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
        "satellite":  "mars.tardigrade.io:7777",
        "bucket":     "my-first-bucket",
        "uploadPath": "foo/bar/baz",
        "encryptionpassphrase": "test"
    }
```

* Store both these files in a `config` folder.  By default, the configurations in `config` will be used unless otherwise specified.

## Run the command-line tool

**NOTE**: The following commands operate in a Linux system

* Get help
```
    $ ./storj_fluree -h
```

* Check version
```
    $ ./storj_fluree -v
```

* Snapshot a Database
This command will create a (local) snapshot for your configured database. The snapshot captures the state and history of a FlureeDB ledger up to the present moment. 
```
    $ ./storj_fluree snapshot 
```

You can optionally specify `debug` when calling this command (by default, `false`) and a filename for database configuration file (by default, `config/db_property.json` is used).

```
    $ ./storj_fluree snapshot debug
```

```
    $ ./storj_fluree snapshot debug ./configuration/db_property.json
```

```
    $ ./storj_fluree snapshot ./configuration/db_property.json debug
```


* List Database Snapshots
This command will list all (local) snapshots for your configured database. 

```
    $ ./storj_fluree list
```

You can optionally specify `debug` when calling this command (by default, `false`) and a filename for database configuration file (by default, `config/db_property.json` is used).

```
    $ ./storj_fluree list debug
```

```
    $ ./storj_fluree list debug ./configuration/db_property.json
```

```
    $ ./storj_fluree list ./configuration/db_property.json debug
```

* Test
This command will read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object.

```
    $ ./storj_fluree test
```

You can optionally specify `debug` when calling this command (by default, `false`) and a filename for the Storj configuration file (by default, `config/storj_property.json` is used).

```
    $ ./storj_fluree test debug
```

```
    $ ./storj_fluree test debug ./configuration/storj_property.json
```

```
    $ ./storj_fluree test ./configuration/storj_property.json debug
```


* Store
This command will read and parse Storj network's configuration, in JSON format, from a desired file and upload your latest snapshot for your configured database. You can optionally specify a specific snapshot. 

```
    $ ./storj_fluree store
```

You can optionally specify `debug` when calling this command (by default, `false`). If you'd like to specify a specific snapshot, you need to specify all three of the following options in order: Fluree config file name, Storj config file name, snapshot name.

If using debug, will attempt to download the data that has been uploaded to Storj.

```
    $ ./storj_fluree store debug
```

```
    $ ./storj_fluree store debug ./configuration/db_property.json ./configuration/storj_config.json 1574091452788.avro
```

```
    $ ./storj_fluree store ./configuration/db_property.json ./configuration/storj_config.json 1574091452788.avro debug
```

