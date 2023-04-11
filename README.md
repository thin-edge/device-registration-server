# device-registration-server
thin-edge.io plugin to register child devices.

## Plugin summary

A http server which provides a simple interface for child devices to register themselves and indicate which operations they support via a HTTP API.

The device-registration-server should be installed on the device where thin-edge.io is running as it needs access to the same filesystem.

### What will be deployed to the device?

* A service called `device-registration-server`

**Technical summary**

The following details the technical aspects of the plugin to get an idea what systems it supports.

|||
|--|--|
|**Languages**|`golang`|
|**CPU Architectures**|`armv7/arm64/x86_64` (though it can be built for other architectures)|
|**Supported init systems**|`systemd` and `init.d/open-rc`|
|**Required Dependencies**|-|
|**Optional Dependencies (feature specific)**|-|

### How to do I get it?

The following linux package formats are provided on the releases page and also in the [tedge-community](https://cloudsmith.io/~thinedge/repos/community/packages/) repository:

|Operating System|Repository link|
|--|--|
|Debian/Raspbian (deb)|[![Latest version of 'device-registration-server' @ Cloudsmith](https://api-prd.cloudsmith.io/v1/badges/version/thinedge/community/deb/device-registration-server/latest/a=all;d=any-distro%252Fany-version;t=binary/?render=true&show_latest=true)](https://cloudsmith.io/~thinedge/repos/community/packages/detail/deb/device-registration-server/latest/a=all;d=any-distro%252Fany-version;t=binary/)|
|Alpine Linux (apk)|[![Latest version of 'device-registration-server' @ Cloudsmith](https://api-prd.cloudsmith.io/v1/badges/version/thinedge/community/alpine/device-registration-server/latest/a=noarch;d=alpine%252Fany-version/?render=true&show_latest=true)](https://cloudsmith.io/~thinedge/repos/community/packages/detail/alpine/device-registration-server/latest/a=noarch;d=alpine%252Fany-version/)|
|RHEL/CentOS/Fedora (rpm)|[![Latest version of 'device-registration-server' @ Cloudsmith](https://api-prd.cloudsmith.io/v1/badges/version/thinedge/community/rpm/device-registration-server/latest/a=noarch;d=any-distro%252Fany-version;t=binary/?render=true&show_latest=true)](https://cloudsmith.io/~thinedge/repos/community/packages/detail/rpm/device-registration-server/latest/a=noarch;d=any-distro%252Fany-version;t=binary/)|


## Documentation

### API

#### POST /register: Register child device

A child device can register itself to the service

**Example**

```sh
curl \
    http://127.0.0.1:9000/register \
    -X POST \
    --data '{"name":"mychild","supportedOperations":["c8y_Firmware"]}' -H "Content-Type: application/json"
```

**Response: 201**

```json
{
    "id": "tedge01_mychild",
    "name": "mychild",
    "parent": "tedge01",
}
```

|Property|Description|
|----|----|
|`id`|Child device id that should be used in all communication with thin-edge.io|
|`name`|Local child device name. This will be the value that was sent by the child device whilst registering itself|
|`parent`|Id of the parent device (for reference only)|

### Configuration

The `device-registration-service` can be controlled by either environment variables or command line flags.

The following flags are supported.

```sh
  --bind string
        Bind address to which the http server should attach to. It listens on all adapters by default.
  --config-dir string
        thin-edge.io base configuration directory (default "/etc/tedge")
  --device-id string
        Use static device id instead of using the tedge cli
  --port int
        Port (default 9000)
  --separator string
        Device name separator (default "_")
  --version
        Show version information
```

Alternatively the values for the flags can be provided via environment variables. The mapping between the flags to environment variables is as follows:

|Flag|Env Variable|
|----|------------|
|`--device-id`|`REGISTRATION_DEVICE_ID`|
|`--config-dir`|`REGISTRATION_CONFIG_DIR`|
|`--bind`|`REGISTRATION_BIND`|
|`--port`|`REGISTRATION_PORT`|
|`--separator`|`REGISTRATION_SEPARATOR`|

### Building

You can build the project and all of the packages by using, though the command will fail if you have uncommitted changes.

```sh
just release
```

If you have uncommitted changes then you can build a snapshot with the following command:

```sh
just release-snapshot
```
