# Building #
1. Run `go build -o installer.exe .\main.go .\utilities.go`
> Note: You can change the build location. Replace **installer.exe** with **Build/installer.exe** (Assets and Run.ps1 must be copied over)
2. Run `go run .\main.go .\utilities.go` or `.\installer.exe`

# Errors #
> 1. CMD errors: Make sure the bridge is not running anywhere else  
> 2. Just closes: Make sure the shortcut points to a valid location (Search bridge, right click and open file location, right click and look at properties).
    (If you decide to change the installation directory then delete downloads and run again)

# Debugging #
1. If it just closes then look at Errors 2 ^
2. If you're unable to see the output because it closes then locate installer.exe, right click and select "Open with Powershell" or the Windows Terminal. (You can use cmd prompt but it doesn't look nice). Anyways, just run it and it should give you an error.

# Uninstalling #
Just delete the directory and the shortcut

Installing
![](https://i.imgur.com/CjUXb6O.png)  
Running 1  
![](https://imgur.com/Lzzr4Zi.png)  
Running 2  
![](https://imgur.com/21q2mNU.png) Nice VSCode