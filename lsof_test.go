package lsof

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadStatus(t *testing.T) {
	s, err := readStatus(os.Getpid())
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NotEmpty(t, s)
	t.Log(s)
}

func TestReadProc(t *testing.T) {
	var stat *Stat
	var err error
	if os.Getenv("GOROOT") != "" {
		stat, err = GetStat(os.Getenv("GOROOT") + "/bin/go")
	} else {
		stat, err = GetStat("/usr/local/go/bin/go")
	}
	require.NoError(t, err)
	require.NotEmpty(t, stat.Ino)
	ps, err := ReadMap(stat.Ino)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.NotEmpty(t, ps)
	t.Log(ps)
}

func TestNotFound(t *testing.T) {
	ps, err := ReadMap(1234)
	require.NoError(t, err)
	require.Nil(t, ps)
}
