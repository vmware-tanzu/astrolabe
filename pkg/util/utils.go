package util

import (
	"github.com/sirupsen/logrus"
	"strings"
)

func GetStringFromParamsMap(params map[string]interface{}, key string, logger logrus.FieldLogger) (value string, ok bool) {
	valueIF, ok := params[key]
	if ok {
		value, ok := valueIF.(string)
		if !ok {
			logger.Errorf("Value for params key %s is not a string", key)
		}
		return value, ok
	} else {
		logger.Errorf("No such key %s in params map", key)
		return "", ok
	}
}

func IsConnectionResetError(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}
	return false
}