
emulator: FORCE
	echo 'Starting Emulator'
	go build -C emulator -v -o emulator cmd/emulator/main.go
	cd emulator; sudo XDG_RUNTIME_DIR="/run" ./emulator
trace:
	docker exec GSElAlamo iperf3 -s 
	docker exec GSElAlamo iperf3 -c GSElAlamo
	docker exec GSKoto
	docker exec GSKoto

clean:
	rm network_emulator

FORCE: ;
