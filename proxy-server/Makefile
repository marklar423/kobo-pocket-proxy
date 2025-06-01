SERVER_EXECUTABLE=pocket-proxy-server

build:
	go build -o ../bin/proxy-server/${SERVER_EXECUTABLE} server.go

run: build
	../bin/proxy-server/${SERVER_EXECUTABLE}

clean:
	go clean
	rm -f ../bin/proxy-server/${SERVER_EXECUTABLE}