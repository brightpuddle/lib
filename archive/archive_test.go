package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	a := assert.New(t)

	// Valid file
	arc, err := Open("testdata/data.zip")
	defer os.RemoveAll(arc.TmpDir)
	a.NoError(err)
	info, err := os.Stat(filepath.Join(arc.TmpDir, "fvTenant.json"))
	if !a.NoError(err) {
		a.True(info.Size() > 0)
	}

	// Invalid file
	_, err = Open("non-existent")
	a.Error(err)
}

func TestClose(t *testing.T) {
	arc, _ := Open("testdata/data.zip")
	err := arc.Close()
	assert.NoError(t, err)
}
