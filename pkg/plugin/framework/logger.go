/*
Copyright the Astrolabe contributors.
Copyright the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"github.com/sirupsen/logrus"
)

// NewLogger returns a logger that is suitable for use within an
// Astrolabe plugin.
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	/*
			!!!DO NOT SET THE OUTPUT TO STDOUT!!!

			go-plugin uses stdout for a communications protocol between client and server.

			stderr is used for log messages from server to client. The astrolabe server makes sure they are logged to the correct
		    destination.
	*/

	// we use the JSON formatter because go-plugin will parse incoming
	// JSON on stderr and use it to create structured log entries.
	logger.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			// this is the hclog-compatible message field
			logrus.FieldKeyMsg: "@message",
		},
		// Astrolabe server already adds timestamps when emitting logs, so
		// don't do it within the plugin.
		DisableTimestamp: true,
	}

	return logger
}
