run : parani
	cd html/sample_images/ && rm -f *.html && ../../parani

GO=go

parani : main.go go.mod makefile html/index.html
	$(GO) version
	printenv | grep PATH
	$(GO) build

parani.exe : main.go go.mod makefile html/index.html
	GOOS=windows GOARCH=amd64 $(GO) build -o parani.exe

clean:
	rm -f parani
	rm -f parani.exe
	rm -f html/sample_images/*.html
