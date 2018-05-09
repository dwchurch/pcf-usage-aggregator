# PCF Usage Aggregator

## Usage

**Choose your path:**
1. Build using go
    - Ensure golang is installed
    - Ensure $GOPATH is set
    - git clone this repo
    - `cd pcf-usage-aggregator`
    - `go build`
    - `./pcf-usage-aggregator`

1. Use the provided binaries on github under releases tab
    - Download binaries [here](https://github.com/Oskoss/pcf-usage-aggregator/releases)

## Configuration

This application expects a `config.yml` in the directory where it will be run. This YAML file will indicate which PCF Foundations to pull data from.

Below is a sample of what is expected in the `config.yml`. **At least one PCF foundation is required.**

**Note:** `admin_password` corresponds to the PCF UAA Admin user of the foundation.

```
foundations:
- name: FoundationName1
  url: sys.pcf.foundationname1.company.com
  admin_password: supersecret1
- name: FoundationName2
  url: sys.pcf.foundationname2.company.com
  admin_password: supersecret2
```

## API

`GET /v1/apps` -> JSON list of all apps on all foundations
