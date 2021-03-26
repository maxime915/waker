waker:		main.go
	go build -o $@

piwaker:	main.go
	env GOOS=linux GOARCH=arm GOARM=5 go build -o $@

.PHONY:
all:	waker | piwaker

.PHONY:
clean:
	rm waker piwaker