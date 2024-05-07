
KERNEL = $(shell uname -r)

.PHONY: build
build: server.c server_xdp.o
	gcc -O2 server.c -o server -lxdp -lbpf -lz -lelf

worker: worker.go
	go build worker.go

server_xdp.o: server_xdp.c worker
	clang -O2 -g -Ilibbpf/src -target bpf -c server_xdp.c -o server_xdp.o

.PHONY: clean
clean:
	rm -f worker
	rm -f server
	rm -f *.o