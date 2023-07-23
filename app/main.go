package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glaslos/lsof"
)

func main() {
	arg := os.Args[1]
	target, err := filepath.Abs(arg)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := lsof.GetStat(target)
	if err != nil {
		log.Fatal(err)
	}

	ps, err := lsof.ReadMap(stat.Ino)
	if err != nil {
		log.Fatal(err)
	}
	if ps == nil {
		os.Exit(1)
	}

	lines := []lsof.Line{
		{
			ProcStatus: *ps,
			Stat:       *stat,
		},
	}

	lsof.Render(lines)
}
