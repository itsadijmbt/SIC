


Write-Output "Building for Windows..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "Adi's SIC-Windows.exe" main.go


Write-Output "Building for Linux..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "Adi's SIC-Linux" main.go


Write-Output "Building for macOS..."
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o "Adi's SIC-macOS" main.go


Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH
