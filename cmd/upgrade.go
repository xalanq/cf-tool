package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/xalanq/cf-tool/util"
)

func less(a, b string) bool {
	reg := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	x := reg.FindSubmatch([]byte(a))
	y := reg.FindSubmatch([]byte(b))
	num := func(s []byte) int {
		n, _ := strconv.Atoi(string(s))
		return n
	}
	for i := 1; i <= 3; i++ {
		if num(x[i]) < num(y[i]) {
			return true
		} else if num(x[i]) > num(y[i]) {
			return false
		}
	}
	return false
}

func getLatest() (version, note, ptime, url string, size uint, err error) {
	goos := ""
	switch runtime.GOOS {
	case "darwin":
		goos = "osx"
	case "linux":
		goos = "linux"
	case "windows":
		goos = "win"
	default:
		err = fmt.Errorf("Not support %v", runtime.GOOS)
		return
	}

	arch := ""
	switch runtime.GOARCH {
	case "386":
		arch = "32"
	case "amd64":
		arch = "64"
	default:
		err = fmt.Errorf("Not support %v", runtime.GOARCH)
		return
	}

	resp, err := http.Get("https://api.github.com/repos/xalanq/cf-tool/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result := make(map[string]interface{})
	json.Unmarshal(body, &result)
	version = result["tag_name"].(string)
	note = result["body"].(string)
	tm, _ := time.Parse("2006-01-02T15:04:05Z", result["published_at"].(string))
	ptime = tm.In(time.Local).Format("2006-01-02 15:04")
	url = fmt.Sprintf("https://github.com/xalanq/cf-tool/releases/download/%v/cf_%v_%v_%v.zip", version, version, goos, arch)
	assets, _ := result["assets"].([]interface{})
	for _, tmp := range assets {
		asset, _ := tmp.(map[string]interface{})
		if url == asset["browser_download_url"] {
			size = uint(asset["size"].(float64))
			break
		}
	}
	return
}

// WriteCounter progress counter
type WriteCounter struct {
	Count, Total uint
	last         uint
}

// Print print progress
func (w *WriteCounter) Print() {
	fmt.Printf("\rProgress: %v/%v KB  Speed: %v KB/s", w.Count/1024, w.Total/1024, (w.Count-w.last)/1024)
	w.last = w.Count
}

func (w *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	w.Count += uint(n)
	return n, nil
}

func upgrade(url, exe string, size uint) (err error) {
	color.Cyan("Download %v", url)
	counter := &WriteCounter{Count: 0, Total: size, last: 0}
	counter.Print()

	ticker := time.NewTicker(time.Second)
	go func() {
		for range ticker.C {
			counter.Print()
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(io.TeeReader(resp.Body, counter))
	if err != nil {
		ticker.Stop()
		counter.Print()
		fmt.Println()
		return
	}
	ticker.Stop()
	counter.Print()
	fmt.Println()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return
	}

	rc, err := reader.File[0].Open()
	if err != nil {
		return
	}
	defer rc.Close()

	newPath := filepath.Join(os.TempDir(), fmt.Sprintf(".%s.new", filepath.Base(exe)))
	oldPath := filepath.Join(os.TempDir(), fmt.Sprintf(".%s.old", filepath.Base(exe)))

	file, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, rc)
	if err != nil {
		return
	}
	file.Close()

	err = os.Rename(exe, oldPath)
	if err != nil {
		os.Remove(newPath)
		return
	}

	err = os.Rename(newPath, exe)
	if err != nil {
		os.Rename(oldPath, exe)
		os.Remove(newPath)
		return
	}

	os.Remove(oldPath)
	return nil
}

// Upgrade itself
func Upgrade(version string) error {
	color.Cyan("Checking version")
	latest, note, ptime, url, size, err := getLatest()
	if err != nil {
		return err
	}
	if !less(version, latest) {
		color.Green("Current version %v is the latest", version)
		return nil
	}

	color.Red("Current version is %v", version)
	color.Green("The latest version is %v, published at %v", latest, ptime)
	fmt.Println(note)

	if !util.YesOrNo("Do you want to upgrade (y/n)? ") {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return err
	}

	if err = upgrade(url, exe, size); err != nil {
		return err
	}

	color.Green("Successfully updated to version %v", latest)
	return nil
}
