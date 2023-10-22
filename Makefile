
emulator: FORCE
	echo "Building emulator"
	go build -C emulator -v -o emulator cmd/emulator/main.go
	echo "Generating satellite positions"
	python satellite_positions.py
	echo 'Starting Emulator'
	cd emulator; sudo ISRAEL=TRUE ./emulator
trace:
	docker exec GSElAlamo iperf3 -s 
	docker exec GSElAlamo iperf3 -c GSElAlamo
	docker exec GSKoto
	docker exec GSKoto

clean:
	rm network_emulator

FORCE: ;
