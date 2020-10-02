package util

import (
	"testing"
    "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckConnectionResetError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Connection reset by peer error should return true",
			err:      errors.New("read tcp 10.244.2.4:38316->10.208.22.141:443: read: connection reset by peer"),
			expected: true,
		},
		{
			name:     "Other error should return false",
			err:      errors.New("This is not a connection reset error"),
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isConnectionError := IsConnectionResetError(test.err)
			assert.Equal(t, test.expected, isConnectionError)
		})
	}
}
