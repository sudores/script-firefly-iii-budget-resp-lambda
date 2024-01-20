build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o main main.go
	zip bootstrap.zip main 
run: build
	./scripts/startup.sh
