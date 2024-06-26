GOOS=windows GOARCH=amd64 go build -o ../template_generator_win.exe

GOOS=linux GOARCH=amd64 go build -o ../template_generator_linux.exe

GOOS=darwin GOARCH=amd64 go build -o ../template_generator_mac_intel.exe

GOOS=darwin GOARCH=arm64 go build -o ../template_generator_mac_apple.exe
