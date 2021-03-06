package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/reusee/e/v2"
)

var (
	me     = e.Default.WithStack()
	ce, he = e.New(me)
	pt     = fmt.Printf
)

func main() {

	// get info
	retry := 10
do:
	resp, err := http.Get("https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=en-US")
	if err != nil {
		if retry > 0 {
			retry--
			goto do
		}
		ce(err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	ce(err)

	var data struct {
		Images []struct {
			URL string
			Hsh string
		}
	}
	if err := json.Unmarshal(content, &data); err != nil {
		if retry > 0 {
			retry--
			goto do
		}
	}
	if len(data.Images) == 0 {
		return
	}

	// get file
	configDir, err := os.UserConfigDir()
	ce(err)
	imageDir := filepath.Join(configDir, "wallpapers")
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		ce(os.Mkdir(imageDir, 0755))
	}
	filePath := filepath.Join(imageDir, data.Images[0].Hsh)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// file not exists, download
		resp, err = http.Get("https://bing.com" + data.Images[0].URL)
		if err != nil {
			if retry > 0 {
				retry--
				goto do
			}
			ce(err)
		}
		defer resp.Body.Close()
		f, err := os.Create(filePath + ".tmp")
		ce(err)
		if _, err := io.Copy(f, resp.Body); err != nil {
			if retry > 0 {
				retry--
				goto do
			}
			ce(err)
		}
		f.Close()
		ce(os.Rename(filePath+".tmp", filePath))

	} else if err != nil {
		ce(err)
	}

	fehPath, err := exec.LookPath("feh")
	if err == nil {
		out, err := exec.Command(fehPath, "--bg-fill", filePath).CombinedOutput()
		pt("%s\n", out)
		_ = err
	}

	out, err := exec.Command(
		"gsettings", "set", "org.gnome.desktop.background", "picture-uri",
		"file://"+filePath,
	).CombinedOutput()
	pt("%s\n", out)
	_ = err

	out, err = exec.Command(
		"osascript",
		"-e",
		`tell application "Finder" to set desktop picture to POSIX file "`+filePath+`"`,
	).CombinedOutput()
	pt("%s\n", out)
	_ = err

}
