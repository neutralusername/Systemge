GOOS=windows GOARCH=amd64 go build -o ../generate_application_template_win.exe

GOOS=linux GOARCH=amd64 go build -o ../generate_application_template_linux.exe

GOOS=darwin GOARCH=amd64 go build -o ../generate_application_template_mac_intel.exe

GOOS=darwin GOARCH=arm64 go build -o ../generate_application_template_mac_apple.exe
