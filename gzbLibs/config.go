package gzbLibs

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
)

const (
	CONFIG_FILE = "./config.json"
)

var (
	config *Config
)

func init() {
	loadConfig()
}

type Config struct {
	MergeCSS bool
	MergeJS  bool
	Version  string
	Ip       string
	Port     uint64
	DataDir  string
}

func loadConfig() {
	rawJSON, err := ioutil.ReadFile(CONFIG_FILE)
	if nil != err {
		log.Fatal("Cannot read config file ", err)
	}
	errJson := json.Unmarshal(rawJSON, &config)
	if nil != errJson {
		log.Fatal("JSON decode error ", errJson)
	}
}

func GetIp() string {
	return config.Ip
}

func GetPort() string {
	return strconv.FormatUint(config.Port, 10)
}

func GetVersion() string {
	return config.Version
}

func GetDataDir() string {
	return config.DataDir
}
