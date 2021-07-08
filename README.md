# dd-prof-upload

dd-prof-upload uploads pprof files to Datadog's Continuous Profiler. This is useful for sharing pprof files you captured yourself, or to reupload downloaded profiles after their retention period expired.

Quick Start:

```
# install
go install github.com/felixge/dd-prof-upload/...@latest

# upload example propfiles
export DD_API_KEY=...
dd-prof-upload ./example/*
```

**Important:** Your profiles should be named exactly like the ones in the [example](./example) folder, e.g. `cpu.pprof` for the CPU profile. Using different names might work, but some features might not work correctly.

Command line options:

```
Usage of dd-prof-upload:
  -env string
    	The name of the environment to assign to the uploaded profiles. (default "dev")
  -key string
    	A Datadog API key for your account. Defaults to DD_API_KEY. (default "")
  -runtime string
    	The name of the runtime to attribute the profiles to. (default "go")
  -service string
    	The name of the service to assign for the uploaded profiles. (default "dd-prof-upload")
  -site string
    	The datadog site to upload to. Defaults to DD_SITE or "datadog.com". (default "datadog.com")
```
