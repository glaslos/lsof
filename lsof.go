package lsof

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type ProcStatus struct {
	PID  int
	PPID int
	Name string
	User *user.User
	UIDs [4]string
	GIDs [4]string
}

func (ps *ProcStatus) String() string {
	str := fmt.Sprintf("Name: %s, PID: %d, PPID: %d", ps.Name, ps.PID, ps.PPID)
	if ps.User != nil {
		str += fmt.Sprintf(", User: %s", ps.User.Username)
	}
	return str
}

func (ps *ProcStatus) read(key, value string, vi uint64) {
	switch key {
	case "Name":
		ps.Name = value
	case "PPid":
		ps.PPID = int(vi)
	case "Uid":
		copy(ps.UIDs[:], strings.Split(value, "\t"))
	case "Gid":
		copy(ps.GIDs[:], strings.Split(value, "\t"))
	}
}

func readProcessStatus(pid int) (*ProcStatus, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return nil, err
	}

	ps := &ProcStatus{PID: pid}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}

		kv := strings.SplitN(line, ":", 2)
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		// v = strings.TrimSuffix(v, " kB")

		// instentionally skipping parse errors
		vi, _ := strconv.ParseUint(v, 10, 64)

		ps.read(k, v, vi)
	}

	return ps, nil
}

type Stat struct {
	Ino  uint64
	Size int64
}

func GetStat(filePath string) (*Stat, error) {
	s := &Stat{}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	stat := fileInfo.Sys().(*syscall.Stat_t)
	s.Ino = stat.Ino
	s.Size = stat.Size
	return s, nil
}

func dirList(dir string) ([]string, error) {
	fh, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	return fh.Readdirnames(-1)
}

func findInode(text string) (uint64, error) {
	fields := strings.Fields(text)
	if len(fields) < 5 {
		return 0, fmt.Errorf("truncated map")
	}

	inode, err := strconv.ParseUint(fields[4], 10, 0)
	if err != nil {
		return 0, err
	}
	return inode, nil
}

func getUser(uid string) (*user.User, error) {
	return user.LookupId(uid)
}

func ReadMap(inode uint64) (*ProcStatus, error) {
	pids, err := dirList("/proc")
	if err != nil {
		return nil, err
	}
	for _, pid := range pids {
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			continue
		}
		ps, err := readProcessStatus(pidInt)
		if err != nil {
			return nil, err
		}

		mdir := "/proc/" + pid + "/maps"

		fh, err := os.Open(mdir)
		if err != nil {
			// permission errors
			continue
		}
		scan := bufio.NewScanner(fh)
		for scan.Scan() {
			fInode, err := findInode(scan.Text())
			if err != nil {
				return nil, err
			}
			if fInode == inode {
				ps.User, err = getUser(ps.UIDs[0])
				if err != nil {
					return nil, err
				}
				return ps, nil
			}
		}
	}
	return nil, nil
}

type Line struct {
	ProcStatus
	Stat
}

func ReadPID(pid int) ([]Line, error) {
	procStat, err := readProcessStatus(pid)
	if err != nil {
		return nil, err
	}
	files, err := dirList("/proc/" + strconv.Itoa(pid) + "/fd")
	if err != nil {
		return nil, err
	}

	lines := []Line{}
	for _, fileName := range files {
		stat, err := GetStat("/proc/" + strconv.Itoa(pid) + "/fd/" + fileName)
		if err != nil || stat == nil {
			continue
		}
		line := Line{
			ProcStatus: *procStat,
		}
		line.Name = fileName
		line.Ino = stat.Ino
		line.Size = stat.Size
		lines = append(lines, line)
	}
	return lines, err
}

func Render(lines []Line) {
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
	for _, line := range lines {
		t.AppendRows([]table.Row{{line.Name, line.PID, "line.User.Username", "txt", "REG", "254,0", line.Size, line.Ino, "arg"}})
	}

	t.Render()
}
