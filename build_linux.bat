set GOOS=linux
set GOARCH=arm64
go build -o forwarding_client_arm64 client/cmd/main.go
go build -o forwarding_server_arm64 server/cmd/main.go