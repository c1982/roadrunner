package roadrunner

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrBuffer_Write_Len(t *testing.T) {
	buf := &errBuffer{buffer: new(bytes.Buffer)}
	buf.Write([]byte("hello"))
	assert.Equal(t, 5, buf.Len())
	assert.Equal(t, buf.String(), "hello")
}
