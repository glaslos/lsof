package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glaslos/lsof"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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

	t := table.NewWriter()
	t.SetStyle(table.Style{
		Box: table.StyleBoxDefault,
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: false,
			SeparateHeader:  false,
			SeparateRows:    false,
			SeparateFooter:  false,
		},
	})
	t.Style().Box.PaddingRight = " "
	t.Style().Box.PaddingLeft = ""
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"COMMAND", "PID", "USER", "FD", "TYPE", "DEVICE", "SIZE/OFF", "NODE", "NAME"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft},
		{Number: 2, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 4, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 5, Align: text.AlignRight, AlignHeader: text.AlignRight, WidthMin: 6},
		{Number: 6, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 7, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 8, Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 9, Align: text.AlignLeft},
	})
	t.AppendRows([]table.Row{{ps.Name, ps.PID, ps.User.Username, "txt", "REG", "254,0", stat.Size, stat.Ino, arg}})
	t.Render()
}
