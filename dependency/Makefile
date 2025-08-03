.PHONY: retrieve test

retrieve:
	@cd retrieval; \
	go build -o retrieve; \
	./retrieve \
	    --buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output); \
	rm retrieve

test:
	@cd test; \
	./test.sh \
		--tarballPath $(tarballPath) \
		--expectedVersion $(version) \
		--expectedOS $(os) \
		--expectedArch $(arch)
