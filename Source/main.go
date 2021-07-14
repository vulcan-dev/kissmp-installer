package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

const url = "https://api.github.com/repos/TheHellBox/KISS-multiplayer/releases/latest"

var log = InitializeLogger()

func InitializeLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		TimestampFormat: "2006-01-02 15:04:05",
		TrimMessages: true,
	})

	return log
}

func main() {
	installerVersion := "1.0.3"
	log.Infoln("Installer made by Vitex#1248")

	git := &Git{}
	git, err := git.GetJSONData("https://api.github.com/repos/vulcan-dev/kissmp-installer/releases/latest"); if err != nil {
		log.Errorln("Something went wrong:", err.Error())
		os.Exit(1)
	}

	if git.Version != installerVersion {
		log.Warnln("[KissMP Installer] New update available")
	}

	if UpdateKissMP() {
		err := DownloadKissMP(); if err != nil {
			log.Errorln("Something went wrong:", err.Error())
		}
	} else {
		err := ListenPipe(); if err != nil {
			if strings.Contains(err.Error(), "101") {
				err = fmt.Errorf("another instance is running")
			}

			log.Errorln("Something went wrong:", err.Error())
		}
	}

	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func UpdateKissMP() bool {
	utilities := Utilities{}
	git := Git{}

	return !utilities.Exists(fmt.Sprintf("./Downloads/Extracted/%s", git.Version))
}

func DownloadKissMP() error {
	git := &Git{}
	git, err := git.GetJSONData(url); if err != nil {
		log.Errorln("Something went wrong:", err.Error())
	}

	filename := git.Assets[0].Name

	log.Infoln("New version available, downloading:", git.Assets[0].Name)

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

			log.Infoln("Successfully Downloaded", filename)
		}
	}

	/* Extract Mod */	
	f := filename[:strings.IndexByte(filename, '.')]
	f = strings.ReplaceAll(f, ".", "_")

	utilities.Unzip(fmt.Sprintf("./Downloads/%s", filename), "./Downloads/Extracted/")
	if err := os.Rename(fmt.Sprintf("./Downloads/Extracted/%s", f), fmt.Sprintf("./Downloads/Extracted/%s", git.Version)); err != nil {
		return err
	}

	/* Remove mod download */
	if utilities.Exists(fmt.Sprintf("./Downloads/%s", filename)) {
		if err := os.Remove(fmt.Sprintf("./Downloads/%s", filename)); err != nil {
			return err
		}
	}

	var gameDirectory string = ""

	key, _ := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\BeamNG\BeamNG.drive\`, registry.QUERY_VALUE)
	defer key.Close()

	val, _, _ := key.GetStringValue("userpath_override")

	if strings.Contains(val, "\\") {
		gameDirectory = val
		log.Infoln("Using Registry Userdata")
	} else if utilities.Exists(fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))) {
		log.Infoln("Using Local Userdata")
		gameDirectory = fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))
	}

	log.Infoln("Game Directory Found:", gameDirectory)

	/* Move the mod */
	tempMod, err := os.Open(fmt.Sprintf("./Downloads/Extracted/%s/KISSMultiplayer.zip", git.Version)); if err != nil {
		return err
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
		return err
	}; defer destination.Close()

	_, err = io.Copy(destination, tempMod); if err != nil {
		return err
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0])); if err != nil {
		return err
    }

	shortcutFile := fmt.Sprintf(`
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

	runFile := fmt.Sprintf(`
	$p = Start-Process "%s/Installer.exe" -ArgumentList "invalidhost" -wait -NoNewWindow -PassThru
	$p.HasExited
	$p.ExitCode
	$startExe = new-object System.Diagnostics.ProcessStartInfo -args PowerShell.exe
	$startExe.verbs
	`, dir)

	utilities.CreateFile(fmt.Sprintf("%s\\shortcut_%s.bat", dir, git.Version), []byte(shortcutFile))
	utilities.CreateFile(fmt.Sprintf("%s\\Run.ps1", dir), []byte(runFile))

	cmd := exec.Command(fmt.Sprintf("%s\\shortcut_%s.bat", dir, git.Version))
	_, err = cmd.Output(); if err != nil {
		return err
	}

	os.Remove(fmt.Sprintf("./shortcut_%s.bat", git.Version))

	log.Infoln("KissMP Bridge Successfully Added to Start Menu")

	return err
}

func ListenPipe() error {
	git := &Git{}
	git, err := git.GetJSONData(url); if err != nil {
		return err
	}

	log.Infoln("Bridge Started (", git.Version, ")")

	cmd := exec.Command(fmt.Sprintf("./Downloads/Extracted/%s/windows/kissmp-bridge.exe", git.Version))
	cmdReader, err := cmd.StdoutPipe(); if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Infoln(scanner.Text(), "\n")
		}
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return err
}