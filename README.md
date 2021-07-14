# Building #
1. Run `go build -o installer.exe .\main.go .\utilities.go`
> Note: You can change the build location. Replace **installer.exe** with **Build/installer.exe** (Assets and Run.ps1 must be copied over)
2. Run `go run .\main.go .\utilities.go` or `.\installer.exe`