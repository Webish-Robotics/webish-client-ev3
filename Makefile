all: bin/serverish
PLATFORM=local
.PHONY: bin/serverish
bin/serverish:
	@docker build . --target bin --output bin/ --platform ${PLATFORM}