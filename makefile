# Binary name
BINARY=redundantFinder
# Builds the project
build:
		go build -o ${BINARY}
		go test -v
# Installs our project: copies binaries
install:
		go install
release:
		# Clean
		go clean
		rm -rf *.gz
		# Build for mac(x64)
		go clean
		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY}_macos_amd64
		# Build for linux(x64)
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${BINARY}_linux_amd64
		# Build for win(x64)
		go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/${BINARY}_windows_amd64.exe
		# Build for mac(386)
		go clean
		CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o bin/${BINARY}_macos_386
		# Build for linux(386)
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o bin/${BINARY}_linux_386
		# Build for win(386)
		go clean
		CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o bin/${BINARY}_windows_386.exe
		# Build for linux(arm)
		go clean
		CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o bin/${BINARY}_linux_arm
		# Build for freebsd(386)
		go clean
		CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -o bin/${BINARY}_freebsd_386
		# Build for freebsd(x64)
		go clean
		CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -o bin/${BINARY}_freebsd_amd64
		go clean
# Cleans our projects: deletes binaries
clean:
		go clean

.PHONY:  clean build
