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
	stat, err := GetStat("/usr/local/go/bin/go")
	require.NoError(t, err)
	require.NotEmpty(t, stat.Ino)
	ps, err := ReadMap(stat.Ino)
	require.NoError(t, err)
	require.NotNil(t, ps)
	require.NotEmpty(t, ps)
	t.Log(ps)
}
