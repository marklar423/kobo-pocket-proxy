SERVER_EXECUTABLE=pocket-proxy-server
MOCK_EXECUTABLE=mock-website

build:
	go build -o ../bin/proxy-server/${SERVER_EXECUTABLE} main.go

build-mock:
	go build -o ../bin/mock-website/${MOCK_EXECUTABLE} mockwebsite/main.go

run: build
	../bin/proxy-server/${SERVER_EXECUTABLE}

run-mock: build-mock
	../bin/mock-website/${MOCK_EXECUTABLE}

clean:
	go clean
	rm -f ../bin/proxy-server/${SERVER_EXECUTABLE}
	rm -f ../bin/mock-website/${MOCK_EXECUTABLE}