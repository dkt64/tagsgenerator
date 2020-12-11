set GOOS=windows
set GOARCH=amd64
go build -o bin/win64/tagsgenerator.exe
set GOOS=windows
set GOARCH=386
go build -o bin/win32/tagsgenerator.exe
set GOOS=linux
set GOARCH=amd64
go build -o bin/lin64/tagsgenerator
set GOOS=linux
set GOARCH=arm64
go build -o bin/arm64/tagsgenerator
set GOOS=darwin
set GOARCH=amd64
go build -o bin/mac64/tagsgenerator