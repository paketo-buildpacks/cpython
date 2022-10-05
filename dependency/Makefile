.PHONY: retrieve

retrieve:
	@cd retrieval; \
	go build -o retrieve; \
	./retrieve \
	    --buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output); \
	rm retrieve
