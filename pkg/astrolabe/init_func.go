package astrolabe

import (
	"github.com/sirupsen/logrus"
)

type InitFunc func(params map[string]interface{},
	s3Config S3Config,
	logger logrus.FieldLogger) (ProtectedEntityTypeManager, error)
