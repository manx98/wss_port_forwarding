set GOOS=windows
set GOARCH=amd64
go build -o forwarding_client_amd64.exe client/cmd/main.go
set GOOS=linux
set GOARCH=arm64
go build -o forwarding_server_arm64 server/cmd/main.go