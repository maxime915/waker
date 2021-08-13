waker:		main.go
	go build -o $@

piwaker:	main.go
	env GOOS=linux GOARCH=arm GOARM=5 go build -o $@

.PHONY: vet
vet:	main.go waker_test.go
	go vet ./...

.PHONY: test
test:	main.go waker_test.go
	go test ./...

.PHONY: all
all:	waker piwaker

.PHONY: clean
clean:
	rm -f waker piwaker