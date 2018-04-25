.PHONY: dist

dist:
	rm -rf dist
	mkdir -p dist/
	mkdir -p dist/images
	mkdir -p dist/db/migrations
	go build -o dist/runner
	cp db/dbconf.yml dist/db/
	cp -R db/migrations dist/migrations
	cp config.example.toml dist/config.toml
	tar czf archive.tgz dist

