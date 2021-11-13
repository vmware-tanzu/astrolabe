package util

import "github.com/vmware-tanzu/astrolabe/pkg/astrolabe"

type ConfigInfo struct {
	PEConfigs map[string]map[string]interface{}
	S3Config  astrolabe.S3Config
}

func NewConfigInfo(peConfigs map[string]map[string]interface{}, s3Config astrolabe.S3Config) ConfigInfo {
	return ConfigInfo{
		PEConfigs: peConfigs,
		S3Config:  s3Config,
	}
}
