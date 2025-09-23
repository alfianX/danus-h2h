buildStandAloneWin:
	go build -o ./bin/danus-h2h/standalone/danus-h2h.exe ./cmd/standalone/main.go

buildStandAloneLin:
	go build -o ./bin/danus-h2h/standalone/danus-h2h ./cmd/standalone/main.go

buildServiceWin:
	go build -o ./bin/danus-h2h/service/danus-h2h.exe ./cmd/service/main.go

buildServiceLin:
	go build -o ./bin/danus-h2h/service/danus-h2h ./cmd/service/main.go