package Command

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionExecute(t *testing.T) {

	// use custom writer to capture output
	storeStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	v := NewVersionCommand("0.0.9")
	err := v.Execute()
	assert.Nil(t, err)

	w.Close()
	out, _ := io.ReadAll(r)
	// restore the stdout
	os.Stdout = storeStdout

	assert.Contains(t, string(out), "0.0.9")
}
