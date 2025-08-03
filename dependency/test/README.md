To test locally:

```shell
# Assume $output_dir is the output from the compilation step, with a tarball and a checksum in it.
# Note that the wildcard is not quoted, to allow globbing

# Passing
$ ./test.sh \
  --tarballPath ${output_dir}/*.tgz \
  --expectedVersion 3.10.7
tarballPath=/tmp/output_dir/python_3.10.7_linux_arm64_jammy_ad0be19c.tgz
expectedVersion=3.10.7
All tests passed!

# Failing
$ /tmp/test/test.sh \
  --tarballPath ${output_dir}/*.tgz \
  --expectedVersion 999.999.999
tarballPath=/tmp/output_dir/python_3.10.7_linux_arm64_jammy_ad0be19c.tgz
expectedVersion=999.999.999
Version 3.10.7 does not match expected version 999.999.999
```