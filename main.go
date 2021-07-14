package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

type Git struct {
	Version string `json:"tag_name"`
	Assets []struct {
		DownloadURL string `json:"browser_download_url"`
		Name string `json:"name"`
	}
}

const url = "https://api.github.com/repos/TheHellBox/KISS-multiplayer/releases/latest"

func main() {
	/* Initialize Logger */
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		TimestampFormat: "2006-01-02 15:04:05",
		TrimMessages: true,
	})
	
	log.Infoln("Installer made by Daniel W (Vitex#1248)")
	
	if Update() {
		Download()
	} else {
		response, err := http.Get(url); if err != nil {
			log.Errorln(err)
		}; defer response.Body.Close()
	
		git := Git{}
		if err = json.NewDecoder(response.Body).Decode(&git); err != nil {
		   log.Errorln(err)
		   os.Exit(1)
		}
		
		log.Infoln("Running with KissMP Version:", git.Version)

		cmd := exec.Command(fmt.Sprintf("./Downloads/Extracted/%s/windows/kissmp-bridge.exe", git.Version))
		cmdReader, err := cmd.StdoutPipe(); if err != nil {
			log.Errorln("Error creating stdoutpipe.", err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				log.Infoln(scanner.Text(), "\n")
			}
		}()

		if err = cmd.Start(); err != nil {
			log.Errorln("Error starting bridge.", err)
			os.Exit(1)
		}

		if err = cmd.Wait(); err != nil {
			log.Errorln("Error waiting for cmd. (make sure this isn't running anywhere else)", err)
			os.Exit(1)
		}
	}
}

func Update() bool {
	utilities := Utilities{}
	git := Git{}
	
	return !utilities.Exists(fmt.Sprintf("./Downloads/Extracted/%s", git.Version))
}

func Download() {
	req, err := http.NewRequest("GET", url, nil); if err != nil {
		log.Errorln(err)
	};
	
	req.Header = http.Header{
		"Host": []string{"api.github.com"},
		"Content-Type": []string{"application/json"},
		"User-Agent": []string{"PostmanRuntime/7.28.0"},
		"Authorization": []string{"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"},
	}
	
	client := &http.Client{}
	response, err := client.Do(req); if err != nil {
		log.Errorln(err)
		os.Exit(1)
	}

	git := Git{}
	if err = json.NewDecoder(response.Body).Decode(&git); err != nil {
	   log.Errorln(err)
	   os.Exit(1)
	}

	filename := git.Assets[0].Name
	
	log.Infoln("New version available:", git.Version)

	/* Download File */
	utilities := Utilities{}

	if _, err := os.Stat(fmt.Sprintf("./Downloads/%s", filename)); err != nil {
		err := utilities.DownloadFile(git.Assets[0].DownloadURL, fmt.Sprintf("./Downloads/%s", filename)); if err != nil {
			if strings.Contains(err.Error(), "path specified") {
				if err = os.MkdirAll(filepath.Dir("./Downloads/"), os.ModePerm); err != nil {
					log.Errorln("Failed creating Downloads directory")
				}
			}

			err = utilities.DownloadFile(git.Assets[0].DownloadURL, fmt.Sprintf("./Downloads/%s", filename)); if err != nil {
				log.Errorln("Failed downloading file.", err)
			}

			log.Infoln("Downloaded", filename)
		}
	}

	/* Extract Mod */	
	f := filename[:strings.IndexByte(filename, '.')]
	f = strings.ReplaceAll(f, ".", "_")

	utilities.Unzip(fmt.Sprintf("./Downloads/%s", filename), "./Downloads/Extracted/")
	if err := os.Rename(fmt.Sprintf("./Downloads/Extracted/%s", f), fmt.Sprintf("./Downloads/Extracted/%s", git.Version)); err != nil {
		log.Errorln("Error renaming dir.", err)
	}

	/* Remove mod download */
	if utilities.Exists(fmt.Sprintf("./Downloads/%s", filename)) {
		os.Remove(fmt.Sprintf("./Downloads/%s", filename))
	}

	var gameDirectory string = ""

	if utilities.Exists(fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))) {
		gameDirectory = fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))
		log.Infoln("Game Directory Found:", gameDirectory)
	}

	/* Move the mod */
	tempMod, err := os.Open(fmt.Sprintf("./Downloads/Extracted/%s/KISSMultiplayer.zip", git.Version)); if err != nil {
		log.Errorln("Failed moving mod.", err)
		os.Exit(1)
	}; defer tempMod.Close()
	
	items, _ := ioutil.ReadDir(gameDirectory)
	var latestVersionStr string = "0"
	var latestVersion float64 = 0
    for _, item := range items {
		ver, _ := strconv.ParseFloat(item.Name(), 64)
		if ver > latestVersion {
			latestVersionStr = fmt.Sprintf("%.2f", ver)
		}
    }
	
	destination, err := os.Create(fmt.Sprintf("%s\\%s\\mods\\KISSMultiplayer.zip", gameDirectory, latestVersionStr)); if err != nil {
		log.Errorln("Failed opening mods directory.", err)
		os.Exit(1)
	}; defer destination.Close()

	_, err = io.Copy(destination, tempMod); if err != nil {
		log.Errorln("Failed copying mods.", err)
		os.Exit(1)
	}
	
	// echo oLink.TargetPath = "%%myPath%%/Downloads/Extracted/%s/windows/kissmp-bridge.exe" >> CreateShortcut.vbs
	
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
		log.Fatal(err)
    }
	
	batchFile := fmt.Sprintf(`
		set "myPath=%%~dp0"
		echo Set oWS = WScript.CreateObject("WScript.Shell") > CreateShortcut.vbs
		echo sLinkFile = "%%AppData%%\\Microsoft\Windows\Start Menu\Programs\KissMP Bridge.lnk" >> CreateShortcut.vbs
		echo Set oLink = oWS.CreateShortcut(sLinkFile) >> CreateShortcut.vbs
		echo oLink.WorkingDirectory = "%s/" >> CreateShortcut.vbs
		echo oLink.TargetPath = "powershell.exe" >> CreateShortcut.vbs
		echo oLink.Arguments = "-ExecutionPolicy bypass %s/Run.ps1" >> CreateShortcut.vbs
		echo oLink.Description = "KissMP Bridge" >> CreateShortcut.vbs
		echo oLink.IconLocation = "%s/Assets/Icon.ico" >> CreateShortcut.vbs
		echo oLink.Save >> CreateShortcut.vbs
		cscript CreateShortcut.vbs
		del CreateShortcut.vbs
	`, dir, dir, dir)
	
	file, err := os.OpenFile(fmt.Sprintf("%s\\shortcut_%s.bat", dir, git.Version), os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0755); if err != nil {
        log.Errorln(err)
		os.Exit(1)
    }

    _, err = file.Write([]byte(batchFile)); if err != nil {
        log.Errorln(err)
		os.Exit(1)
    }; file.Close()

	cmd := exec.Command(fmt.Sprintf("%s\\shortcut_%s.bat", dir, git.Version))
	_, err = cmd.Output(); if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}
	
	os.Remove(fmt.Sprintf("./shortcut_%s.bat", git.Version))

	log.Infoln(fmt.Sprintf("KissMP %s Successfully Installed", git.Version))
}

func Uninstall() {
	
}