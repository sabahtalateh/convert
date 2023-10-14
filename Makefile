imports:
	importblocks ./...

test:
	go clean -testcache
	go test ./tests/...

richtest: richgo
	go clean -testcache
	richgo test -v ./tests/...

richgo:
	@if [ -z $(RICHGO) ]; then go install github.com/kyoh86/richgo@latest; fi

