package main

import (
	"os"
	"path"
	"time"
	"log"
	"syscall"
    "io/ioutil"
    "strings"
    "os/exec"
)

var (
	gSavePageFn	    string
    gSavePageDoneFn string
	gDnDir		    string  
	gLockFn		    string
    gNewPageFn      string
    gNewDownFn      string
)

func inits() {
	homeDir := os.Getenv("HOME")
	gLockFn = path.Join(homeDir, "temp", "webber.lock")
	gSavePageFn = path.Join(homeDir, "temp", "savepage.txt")
    gSavePageDoneFn = path.Join(homeDir, "temp", "savepagedone")
	gDnDir = path.Join(homeDir, "Downloads")
    gNewPageFn = path.Join(homeDir, "temp", os.Args[2] + ".page")
    gNewDownFn = path.Join(homeDir, "temp", os.Args[2] + ".down")
}

func launchChrome(url string) *os.Process {
    proc, err := os.StartProcess("/opt/google/chrome/chrome",
        []string{"/opt/google/chrome/chrome", "--new-window", url},
        &os.ProcAttr{})
    if err != nil {
        log.Panic(err)
    }
    return proc    
}

func lock() *os.File {
	f, err := os.OpenFile(gLockFn, os.O_CREATE+os.O_APPEND, 0666)
	if err != nil {
		log.Panic(err)
	}
	fd := f.Fd()
	for {
		err = syscall.Flock(int(fd), syscall.LOCK_EX+syscall.LOCK_NB)
		if err == nil { break }
		time.Sleep(1 * time.Second)
	}
	return f		
}

func findNewFile(i0 []os.FileInfo, i1 []os.FileInfo) string {
    if len(i0) == len(i1) {
        return ""
    }

    for _, j1 := range i1 {
        var found = false
        for _, j0 := range i0 {
            if j1.Name() == j0.Name() {
                found = true
                break
            }
        }
        if !found &&
	        !strings.HasPrefix(j1.Name(), ".com.google.Chrome.") &&
	        !strings.HasSuffix(j1.Name(), ".crdownload") {
            return j1.Name()
        }
    }
    return ""
}

func download(finfo []os.FileInfo) bool {
    finfo1, _ := ioutil.ReadDir(gDnDir)
    n := findNewFile(finfo, finfo1)
    if n != "" {
        movefile(path.Join(gDnDir, n), gNewDownFn)
        return true
    } else {
        return false
    }
}

func movefile(s string, d string) {
    exec.Command("mv", s, d).Run()
}

func savepage() bool {
    _, err := os.Stat(gSavePageDoneFn);
    if err == nil {
        movefile(gSavePageFn, gNewPageFn)
        return true
    } else {
        return false
    }
}

func main() {
	if len(os.Args) < 3 {
		log.Panic(os.Args[0], "url save-prefix")
	}

	inits()

	f := lock()
	defer f.Close()

    os.Remove(gSavePageFn)
    os.Remove(gSavePageDoneFn)
    os.Remove(gNewPageFn)
    os.Remove(gNewDownFn)

    finfo, _ := ioutil.ReadDir(gDnDir)
	proc := launchChrome(os.Args[1])
    for {
        if download(finfo) || savepage() {
            break
        }
        time.Sleep(1*time.Second)
    }
    proc.Kill()
}
