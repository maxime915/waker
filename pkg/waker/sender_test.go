package waker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendErrors(t *testing.T) {
	err := SendPacketTo("", "")
	assert.Error(t, err)

	err = SendPacketTo("", "0.0.0.0:9000")
	assert.Error(t, err)

	err = SendPacketTo("00:00:00:00:00:00", "")
	assert.Error(t, err)

	err = SendPacketTo("invalid-target", "0:9000")
	assert.Error(t, err)

	// don't use true broadcast : the network may be a real one
	err = SendPacketTo("00:00:00:00:00:00", "127.0.0.1:9000")
	assert.NoError(t, err)
}

func TestNewSender(t *testing.T) {
	// _, err := NewOneTimeSender("", "")
	// assert.Error(t, err)

	// _, err = NewOneTimeSender("", "0.0.0.0:9000")
	// assert.Error(t, err)

	// _, err = NewOneTimeSender("00:00:00:00:00:00", "")
	// assert.Error(t, err)

	// _, err = NewOneTimeSender("00:00:00:00:00:00", "0.0.0.0:9000")
	// assert.NoError(t, err)
}
