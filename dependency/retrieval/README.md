# Dependency Retrieval

## Running locally

Run the following command:

```
go run main.go \
  --buildpack-toml-path ../../buildpack.toml \
  --output /path/to/retrieved.json
```

Example output (abbreviated for clarity):

```
Found 175 versions of python from upstream
[
  "3.11.0", "3.10.8", [...],  "2.0.1"
]
Found 1 version of python for constraint 3.11.*
[
  "3.11.0"
]
Found 9 versions of python for constraint 3.10.*
[
  "3.10.8", [...], "3.10.0"
]
Found 15 versions of python for constraint 3.9.*
[
  "3.9.15", [...],  "3.9.0"
]
Found 16 versions of python for constraint 3.8.*
[
  "3.8.15", [...], "3.8.0"
]
Found 16 versions of python for constraint 3.7.*
[
  "3.7.15", [...], "3.7.0"
]
Found 0 versions of python as new versions
[
  
]
Wrote metadata to /path/to/retrieved.json

```