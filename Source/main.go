//go:generate goversioninfo -icon=../Assets/icon.ico

// cd source
// go generate
// go build -o ./Installer.exe

// TODO Figure out how I added the icon
// TODO Auto-launch BeamNG if process doesn't exist

package main

import (
	"bufio"
	"errors"
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
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
)

var (
	log = InitializeLogger()
	git = &Git{}
	utilities = &Utilities{}
)

const INSTALLER_VERSION = "1.1.6"

func main() {
	// setup close handler so you don't get a bad exit code on interrupt
	utilities.SetupCloseHandler()
	
	url := "https://api.github.com/repos/TheHellBox/KISS-multiplayer/releases/latest"
	log.Infoln(fmt.Sprintf(`Installer made by Vitex#1248
	Version: %s
	`, INSTALLER_VERSION))

	git, _ = git.GetJSONData(url)
	git, err := git.GetJSONData("https://api.github.com/repos/vulcan-dev/kissmp-installer/releases/latest"); if err != nil {
		log.Errorln("Something went wrong:", err.Error())
		os.Exit(1)
	}

	// check if there's a new version of the installer
	if git.Version != INSTALLER_VERSION && git.Version != "" {
		log.Warnln("[KissMP Installer] New update available")
		sc := bufio.NewScanner(strings.NewReader(git.Body))
		for sc.Scan() {
			fmt.Println("\t-> " + sc.Text())
		}
	}

	// this can fail because you can be rate limited by the api
	if git.Version == "" {
		log.Fatalln("api.github.com has limited you for sending too many requests, don't spam open the bridge. please try again in an 30-40 minutes")
	}

	_, err = git.GetJSONData(url); if err != nil {
		log.Errorln(err.Error())
	}

	if UpdateKissMP() {
		err := DownloadKissMP(); if err != nil {
			log.Errorln(err.Error())
		}
	} else {
		err := ListenPipe(); if err != nil {
			log.Errorln(err.Error())
		}
	}

	log.Warnln("Press 'Enter' to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func UpdateKissMP() bool {
	return !utilities.Exists(fmt.Sprintf("./Downloads/Extracted/%s", git.Version))
}

func DownloadKissMP() error {
	filename := git.Assets[0].Name
	log.Infoln("New version available, downloading:", git.Assets[0].Name)

	// download the mod from github
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

	// extract the mod
	f := filename[:strings.IndexByte(filename, '.')]
	f = strings.ReplaceAll(f, ".", "_")

	utilities.Unzip(fmt.Sprintf("./Downloads/%s", filename), "./Downloads/Extracted/")
	if err := os.Rename(fmt.Sprintf("./Downloads/Extracted/%s", f), fmt.Sprintf("./Downloads/Extracted/%s", git.Version)); err != nil {
		return errors.New("failed renaming mod file" + ". error: " + err.Error())
	}

	// delete the mod after extracting
	if utilities.Exists(fmt.Sprintf("./Downloads/%s", filename)) {
		if err := os.Remove(fmt.Sprintf("./Downloads/%s", filename)); err != nil {
			return errors.New("failed deleting mod file" + ". error: " + err.Error())
		}
	}

	var gameDirectory string = ""

	// check if beamng's userpath has been overridden
	key, _ := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\BeamNG\BeamNG.drive\`, registry.QUERY_VALUE)
	defer key.Close()

	val, _, _ := key.GetStringValue("userpath_override")

	// if the registry value has a backslash then that indicates that a path was set so use this for the game directory
	if strings.Contains(val, "\\") {
		log.Infoln("Using Registry Userdata")
		gameDirectory = val
	} else if utilities.Exists(fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))) { // path has not been overridden so use the default
		log.Infoln("Using Local Userdata")
		gameDirectory = fmt.Sprintf("%s\\BeamNG.drive", os.Getenv("LocalAppData"))
	}

	log.Infoln("Game Directory Found:", gameDirectory)

	// check if the mod exists so we can copy over it to the mods directory later on
	tempMod, err := os.Open(fmt.Sprintf("./Downloads/Extracted/%s/KISSMultiplayer.zip", git.Version)); if err != nil {
		return errors.New("failed opening KISSMultiplayer.zip" + ". error: " + err.Error())
	}; defer tempMod.Close()

	items, _ := ioutil.ReadDir(gameDirectory)

	var latestVersionStr string = "0"
	var latestVersion float64 = 0

	// loop over each file in the game directory, if it's a number (0.x) then check if that is greater than our version
    for _, item := range items {
		ver, _ := strconv.ParseFloat(item.Name(), 64)

		if ver > latestVersion {
			latestVersionStr = fmt.Sprintf("%.2f", ver)
		}
    }

	// move the downloaded mod to beamng.drive mods directory
	destination, err := os.Create(fmt.Sprintf("%s\\%s\\mods\\KISSMultiplayer.zip", gameDirectory, latestVersionStr)); if err != nil {
		return errors.New("failed creating KISSMultiplayer.zip" + ". error: " + err.Error())
	}; defer destination.Close()

	_, err = io.Copy(destination, tempMod); if err != nil {
		return errors.New("failed copying mod file" + ". error: " + err.Error())
	}

	// get current working directory for use with the powershell & batch scripts
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])); if err != nil {
		return errors.New("failed getting current working directory" + ". error: " + err.Error())
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
		return errors.New("failed getting output for cmd.Output()" + ". error: " + err.Error())
	}

	os.Remove(fmt.Sprintf("./shortcut_%s.bat", git.Version))
	log.Infoln("KissMP Bridge Successfully Added to Start Menu")

	return nil
}

func GetProcessID(name string) (int, error) {
	// find the process id of the process with the given name
	var pid int = 0
	var err error = nil

	processes, err := process.Processes()
	if err != nil {
		return 0, errors.New("failed getting processes" + ". error: " + err.Error())
	}

	for _, proc := range processes {
		n, _ := proc.Name();
		if n == name {
			pid = int(proc.Pid)
		}
	}

	return pid, nil
}

func ListenPipe() error {
	log.Infoln("KissMP Version:", git.Version)
	
	pid, _ := GetProcessID("BeamNG.drive.exe");
	if pid == 0 {
		log.Infoln("Launching Game")
		open.Start("steam://rungameid/284160")
	}

	// execute the kissmp bridge so we can pipe the stdout data to here
	cmd := exec.Command(fmt.Sprintf("./Downloads/Extracted/%s/windows/kissmp-bridge.exe", git.Version))
	cmdReader, err := cmd.StdoutPipe(); if err != nil {
		return errors.New("failed reading stdout pipe. another instance may be running" + ". error: " + err.Error())
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Infoln(scanner.Text(), "\n")
		}
	}()

	if err = cmd.Start(); err != nil {
		return errors.New("cmd.Start() failed in f: ListenPipe" + ". error: " + err.Error())
	}

	if err = cmd.Wait(); err != nil {
		return errors.New("cmd.Wait() failed in f: ListenPipe" + ". error: " + err.Error())
	}

	return nil
}

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