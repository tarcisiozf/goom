run: build
	./goom doom1.wad

build:
	go build -o goom cmd/main.go

pprof:
	go tool pprof -http=:8080 cpu.pprof