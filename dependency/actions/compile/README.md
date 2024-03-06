Running compilation locally:

1. Build the build environment:
```shell
docker build --tag compilation-<target> --file <target>.Dockerfile .

# Noble example
docker build --tag compilation-noble --file noble.Dockerfile .

# Jammy example
docker build --tag compilation-jammy --file jammy.Dockerfile .

# Bionic example
docker build --tag compilation-bionic --file bionic.Dockerfile .
```

2. Make the output directory:
```shell
output_dir=$(mktemp -d)
```

3. Run compilation and use a volume mount to access it:
```shell
$ docker run --volume $output_dir:/tmp/compilation compilation-<target> --outputDir /tmp/compilation --target <target> --version <version> 

# Noble example
$ docker run --volume $output_dir:/tmp/compilation compilation-noble --outputDir /tmp/compilation --target noble --version 3.10.7

# Jammy example
$ docker run --volume $output_dir:/tmp/compilation compilation-jammy --outputDir /tmp/compilation --target jammy --version 3.10.7

# Bionic example
$ docker run --volume $output_dir:/tmp/compilation compilation-bionic --outputDir /tmp/compilation --target bionic --version 3.10.7
```
