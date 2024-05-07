
KERNEL = $(shell uname -r)

.PHONY: build
build: player_server.c player_server_xdp.o
	gcc -O2 player_server.c -o player_server -lxdp -lbpf -lz -lelf

player_server_worker: player_server_worker.go
	go build player_server_worker.go

player_server_xdp.o: player_server_xdp.c player_server_worker
	clang -O2 -g -Ilibbpf/src -target bpf -c player_server_xdp.c -o player_server_xdp.o

.PHONY: clean
clean:
	rm -f client
	rm -f player_server_worker
	rm -f player_server
	rm -f *.o
