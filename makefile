run : parani
	cd html/sample_images/ && rm *.html && ../../parani

parani : main.go go.mod makefile html/index.html
	$(shell goenv which go) build
