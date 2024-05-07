
.PHONY: build
build: server.c worker
	gcc -O2 server.c -o server

worker: worker.go
	go build worker.go

.PHONY: clean
clean:
	rm -f worker
	rm -f server
	rm -f *.o
