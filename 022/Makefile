
KERNEL = $(shell uname -r)

ifeq ($(UNAME), Linux)

.PHONY: build
build: player_server.c player_server_xdp.o zone_database client
	gcc -O2 player_server.c -o player_server -lxdp -lbpf -lz -lelf

player_server_worker: player_server_worker.go zone_database
	go build player_server_worker.go packets.go world.go

player_server_xdp.o: player_server_xdp.c player_server_worker
	clang -O2 -g -Ilibbpf/src -target bpf -c player_server_xdp.c -o player_server_xdp.o

else

.PHONY: build *.go
build: client player zone_database world_server

endif

player: player.go
	go build player.go packets.go world.go

client: client.go
	go build client.go

zone_database: zone_database.go
	go build zone_database.go packets.go world.go

world_server: world_server.go
	go build world_server.go packets.go world.go

.PHONY: test
test: packets.go world.go world_test.go
	go test packets.go world.go world_test.go

.PHONY: clean
clean:
	rm -f client
	rm -f player
	rm -f player_server_worker
	rm -f player_server
	rm -f world_server
	rm -f zone_database
	rm -f *.o
