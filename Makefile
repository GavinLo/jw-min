
OutName=jw-min
GO=go
SrcFile=jw.min.go
OutDir=./bin/

compiler_jar=compiler.jar
htmlcompressor_jar=htmlcompressor-1.5.3.jar
yuicompressor_jar=yuicompressor-2.4.8.jar

build: pre_build $(OutDir)$(OutName)

pre_build :
	@mkdir -p $(OutDir)

$(OutDir)$(OutName) : $(SrcFile)
	$(GO) build -o $@ $<

.PHONY: install
install :
	@cp -f $(OutDir)$(OutName) /usr/local/bin/$(OutName)
	@cp -f $(compiler_jar) /usr/local/bin/$(compiler_jar)
	@cp -f $(htmlcompressor_jar) /usr/local/bin/$(htmlcompressor_jar)
	@cp -f $(yuicompressor_jar) /usr/local/bin/$(yuicompressor_jar)

.PHONY: clean
clean :
	@rm -rf $(OutDir)

