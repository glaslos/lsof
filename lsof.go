package lsof

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type ProcStatus struct {
	PID  int
	PPID int
	Name string
	User *user.User
	UIDs [4]string
	GIDs [4]string
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

func readStatus(pid int) (*ProcStatus, error) {
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
		return s, err
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
		ps, err := readStatus(pidInt)
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
