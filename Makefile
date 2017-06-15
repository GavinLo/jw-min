
OutName=jw-min
GO=go
SrcFile=jw.min.go
OutDir=./bin/

build: pre_build $(OutDir)$(OutName)

pre_build :
	@mkdir -p $(OutDir)

$(OutDir)$(OutName) : $(SrcFile)
	$(GO) build -o $@ $<

.PHONY: install
install :
	@cp $(OutDir)$(OutName) /usr/local/bin/$(OutName)

.PHONY: clean
clean :
	@rm -rf $(OutDir)

