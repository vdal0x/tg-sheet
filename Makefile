APP      = TGSheet
BINARY   = tg-sheet
BUNDLE   = dist/$(APP).app
DMG      = dist/$(APP).dmg
STAGING  = dist/dmg-staging

.PHONY: build bundle dmg clean

build:
	mkdir -p dist
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -o dist/$(BINARY) ./cmd/tg-sheet/

bundle: build
	rm -rf $(BUNDLE)
	mkdir -p $(BUNDLE)/Contents/MacOS
	mkdir -p $(BUNDLE)/Contents/Resources
	cp dist/$(BINARY)     $(BUNDLE)/Contents/MacOS/$(BINARY)
	cp build/Info.plist   $(BUNDLE)/Contents/Info.plist
	cp .env               $(BUNDLE)/Contents/Resources/.env

dmg: bundle
	rm -rf $(STAGING) $(DMG)
	mkdir -p $(STAGING)
	cp -r $(BUNDLE) $(STAGING)/
	# Symlink to /Applications so the user can drag-install
	ln -s /Applications $(STAGING)/Applications
	hdiutil create \
		-volname $(APP) \
		-srcfolder $(STAGING) \
		-ov -format UDZO \
		-o $(DMG)
	rm -rf $(STAGING)
	@echo "→ $(DMG) ready"

clean:
	rm -rf dist/
