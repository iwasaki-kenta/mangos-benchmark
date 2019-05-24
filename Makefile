single:
	env GOOS=linux GOARCH=amd64 go build -o main single.go
	rsync -avz main root@$(ENDPOINT):/root

concurrent:
	env GOOS=linux GOARCH=amd64 go build -o main concurrent.go
	rsync -avz main root@$(ENDPOINT):/root