# Building #
1. Run `go build -o installer.exe .\main.go .\utilities.go`
> Note: You can change the build location. Replace **installer.exe** with **Build/installer.exe** (Assets and Run.ps1 must be copied over)
2. Run `go run .\main.go .\utilities.go` or `.\installer.exe`

# Errors #
> 1. CMD errors: Make sure the bridge is not running anywhere else  
> 2. Just closes: Make sure the shortcut points to a valid location (Search bridge, right click and open file location, right click and look at properties).
    (If you decide to change the installation directory then delete downloads and run again)

# Uninstalling #
Just delete the directory and the shortcut