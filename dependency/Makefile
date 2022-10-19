.PHONY: retrieve test

retrieve:
	@cd retrieval; \
	go build -o retrieve; \
	./retrieve \
	    --buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output); \
	rm retrieve

test:
	./test/test.sh \
		--tarballPath $(tarballPath) \
		--expectedVersion $(version)
