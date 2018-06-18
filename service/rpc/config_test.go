package rpc

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestConfig_Listener(t *testing.T) {
	cfg := &config{Listen: "tcp://:18001"}

	ln, err := cfg.listener()
	assert.NoError(t, err)
	assert.NotNil(t, ln)
	defer ln.Close()

	assert.Equal(t, "tcp", ln.Addr().Network())
	assert.Equal(t, "[::]:18001", ln.Addr().String())
}

func TestConfig_ListenerUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &config{Listen: "unix://rpc.sock"}

	ln, err := cfg.listener()
	assert.NoError(t, err)
	assert.NotNil(t, ln)
	defer ln.Close()

	assert.Equal(t, "unix", ln.Addr().Network())
	assert.Equal(t, "rpc.sock", ln.Addr().String())
}

func Test_Config_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &config{Listen: "uni:unix.sock"}
	ln, err := cfg.listener()
	assert.Nil(t, ln)
	assert.Error(t, err)
	assert.Equal(t, "invalid socket DSN (tcp://:6001, unix://rpc.sock)", err.Error())
}

func Test_Config_ErrorMethod(t *testing.T) {
	cfg := &config{Listen: "xinu://unix.sock"}

	ln, err := cfg.listener()
	assert.Nil(t, ln)
	assert.Error(t, err)
}

func TestConfig_Dialer(t *testing.T) {
	cfg := &config{Listen: "tcp://:18001"}

	ln, err := cfg.listener()
	defer ln.Close()

	conn, err := cfg.dialer()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	assert.Equal(t, "tcp", conn.RemoteAddr().Network())
	assert.Equal(t, "127.0.0.1:18001", conn.RemoteAddr().String())
}

func TestConfig_DialerUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &config{Listen: "unix://rpc.sock"}

	ln, err := cfg.listener()
	defer ln.Close()

	conn, err := cfg.dialer()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	assert.Equal(t, "unix", conn.RemoteAddr().Network())
	assert.Equal(t, "rpc.sock", conn.RemoteAddr().String())
}

func Test_Config_DialerError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &config{Listen: "uni:unix.sock"}
	ln, err := cfg.dialer()
	assert.Nil(t, ln)
	assert.Error(t, err)
	assert.Equal(t, "invalid socket DSN (tcp://:6001, unix://rpc.sock)", err.Error())
}

func Test_Config_DialerErrorMethod(t *testing.T) {
	cfg := &config{Listen: "xinu://unix.sock"}

	ln, err := cfg.dialer()
	assert.Nil(t, ln)
	assert.Error(t, err)
}
