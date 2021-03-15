set GOOS=windows
set GOARCH=amd64
go build -o bin/tagsgenerator-win64/tagsgenerator-win64.exe
set GOOS=windows
set GOARCH=386
go build -o bin/tagsgenerator-win32/tagsgenerator-win32.exe
set GOOS=linux
set GOARCH=amd64
go build -o bin/tagsgenerator-linux_amd64/tagsgenerator-linux-amd64
chmod 0777 bin/tagsgenerator-linux_amd64/tagsgenerator-linux-amd64
set GOOS=linux
set GOARCH=arm64
go build -o bin/tagsgenerator-linux_arm64/tagsgenerator-linux-arm64
chmod 0777 bin/tagsgenerator-linux_arm64/tagsgenerator-linux-arm64
set GOOS=darwin
set GOARCH=amd64
go build -o bin/tagsgenerator-mac_amd64/tagsgenerator-mac-amd64
set GOOS=darwin
set GOARCH=arm64
go build -o bin/tagsgenerator-mac_arm64/tagsgenerator-mac-arm64