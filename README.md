# drone-nunit

## Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/nunit -f docker/Dockerfile .
```

## Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e PLUGIN_TEST_REPORT_PATH=/drone/src/*.xml \
  -e PLUGIN_FAIL_IF_NO_RESULTS=true \
  -e PLUGIN_FAILED_TESTS_FAIL_BUILD=true \
  -e PLUGIN_LOG_LEVEL=debug \
  -w /drone/src \
  -v $(pwd):/drone/src \
  plugins/nunit
```

## Plugin Settings
- `LOG_LEVEL` debug/info Level defines the plugin log level. Set this to debug to see the response from NUnit
- `PLUGIN_TEST_REPORT_PATH` The pattern to find the NUnit xml files. It could support globs, e.g: TestResults/*.xml
- `PLUGIN_FAIL_IF_NO_RESULTS` if set to true, it will fail the build if no tests result files are found for the given pattern
- `PLUGIN_FAILED_TESTS_FAIL_BUILD` if set to true, it will fail the build if there is any failed test

## Supporting Architectures:
This plugin currently supports only Linux on the amd64 architecture. Other architectures and operating systems are not supported at this time.

## Why Debian instead of Alpine?
We initially tried using Alpine for our Docker image due to its lightweight nature. However, Alpine uses musl libc instead of the more common glibc, which caused compatibility issues with some of the libraries our application depends on (like libxml2 and libxslt). These libraries are more stable and compatible in Debian, which uses glibc. To avoid these issues, we've switched to a Debian-based image for better compatibility and stability.
	
