package util

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const peConfigFileSuffix = ".pe.json"

func ReadConfigFiles(confDirPath string) (ConfigInfo, error) {
	configMap := make(map[string]map[string]interface{})

	confDir, err := os.Stat(confDirPath)
	if err != nil {
		log.Panicln("Could not stat configuration directory " + confDirPath)
	}
	if !confDir.Mode().IsDir() {
		log.Panicln(confDirPath + " is not a directory")

	}

	peDirPath := confDirPath + "/pes"
	peDir, err := os.Stat(peDirPath)
	if err != nil {
		log.Panicf("Could not stat protected entity configuration directory %s, err = %v\n", peDirPath, err)
	}
	if !peDir.Mode().IsDir() {
		log.Panicf("%s is not a directory", peDirPath)

	}
	files, err := ioutil.ReadDir(peDirPath)
	if err != nil {
		log.Panicf("Could not list protected entity configuration directory %s, err = %v\n", peDirPath, err)
	}
	for _, curFile := range files {
		if !strings.HasPrefix(curFile.Name(), ".") && strings.HasSuffix(curFile.Name(), peConfigFileSuffix) {
			peTypeName := strings.TrimSuffix(curFile.Name(), peConfigFileSuffix)
			peConf, err := readConfigFile(filepath.Join(peDirPath, curFile.Name()))
			if err != nil {
				log.Panicln("Could not process conf file " + curFile.Name() + " continuing, err = " + err.Error())
			} else {
				configMap[peTypeName] = peConf
			}
		}
	}

	s3ConfFilePath := filepath.Join(confDirPath, "s3config.json")

	s3Config, err := readS3ConfigFile(s3ConfFilePath)
	if err != nil {
		log.Panicf("Could not read S3 configuration directory %s, err = %v\n", s3ConfFilePath, err)
	}

	return NewConfigInfo(configMap, *s3Config), nil
}

func readConfigFile(confFile string) (map[string]interface{}, error) {
	jsonFile, err := os.Open(confFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open conf file %s", confFile)
	}
	defer jsonFile.Close()
	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read conf file %s", confFile)
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonBytes), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal JSON from %s", confFile)
	}
	return result, nil
}

func readS3ConfigFile(s3ConfFile string) (*astrolabe.S3Config, error) {
	jsonFile, err := os.Open(s3ConfFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open conf file %s", s3ConfFile)
	}
	defer jsonFile.Close()
	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read conf file %s", s3ConfFile)
	}
	var result astrolabe.S3Config
	err = json.Unmarshal([]byte(jsonBytes), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal JSON from %s", s3ConfFile)
	}
	return &result, nil
}
