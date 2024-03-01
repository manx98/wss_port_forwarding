set GOOS=windows
set GOARCH=amd64
go build -o forwarding_client_windows_amd64.exe client/cmd/main.go
go build -o forwarding_server_windows_amd64.exe server/cmd/main.go
set GOOS=linux
set GOARCH=arm64
go build -o forwarding_client_linux_arm64 client/cmd/main.go
go build -o forwarding_server_linux_arm64 server/cmd/main.go