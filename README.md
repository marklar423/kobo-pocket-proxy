# Kobo Pocket Proxy

In light of the [Pocket shutdown](https://support.mozilla.org/en-US/kb/future-of-pocket), this project aims to preserve the wonderful Pocket integration on Kobo eReaders. 

This project has two main components:

## 1 - Kobo Mod
This is a modification to be installed on your physical Kobo eReader. It replaces all HTTP API calls to `getpocket.com` or `text.getpocket.com` to instead point to the URL of your choice, as specified by the config file, located at `/.adds/pocket_proxy/pocket_proxy.conf` on your device.

### Installation & Uninstallation
> [!WARNING]  
> This risks bricking your Kobo. I have only tested this on my Kobo Libra 2 - you've been warned.
> If you do brick your device, [Kobo has some instructions](https://help.kobo.com/hc/en-us/articles/360017765713-Manual-reset-your-Kobo-eReader) on how to do a hard reset, which might rescue it (at the cost of wiping all data).

To install copy the KoboRoot.tgz file to the `/.kobo/` directory on your Kobo ereader, and unplug your device. It will install the modification and reboot. You might have to enable hidden folders to see this directory.

To remove the mod, simply delete the file located at `/.adds/pocket_proxy/DELETE_ME_TO_UNINSTALL` on your device, and reboot the device.

### Configuration
To set the URL endpoint that your Kobo will make Pocket requests to, edit `/.adds/pocket_proxy/pocket_proxy.conf` with the following:

```ini
[PocketProxy]
GetSendApiHostPort=http://mypocketproxy.com/
TextApiHostPort=http://mypocketproxy.com/
```

After changing the config file, you might need to reboot your Kobo again.

### How It Works
At startup, Kobo devices dynamically load all libraries in an `/imageformats` directory on the device. This mod is placed there, and when loaded uses [NickelHook](https://github.com/pgaskin/NickelHook) to replace a function called `WebRequester::makeRequest()` with a custom version. The custom function simply replaces outgoing Pocket HTTP calls with the URL set in the config file.

## 2 - Proxy Server

If your chosen Pocket replacement already has a Pocket-compatible API, then simply edit the mod config file to point to the replacement's API. 

If not, then the included proxy server is designed to translate Pocket API calls to the API of one of the supported services, so the Kobo can keep talking to the new backend even after the Pocket API officially shuts down.

Currently, only Readeck is supported, but please feel free to contribute code for other backends.

### Installation & Configuration
Once you have your Readeck instance [running](https://readeck.org/en/start), follow these steps to generate your bearer token:
1. Log into your Readeck instance.
1. Go to Settings > API Tokens
   - (This should be located at `http://myreadeckinstance.com/profile/tokens`)
1. Click "Create a new API token" to create one.
   - Make sure you grant this token Bookmark read & write permissions 
1. Copy the token in the "Your API token" field.

Now, you can start the proxy container, configured to use Readeck as the backend:

```sh
$ podman run kobo-pocket-proxy --backend_endpoint=http://myreadeckinstance.com --backend_bearer_token=123
```

Or if running the binary directly:

```sh
$ pocket-proxy-server --backend_endpoint=http://myreadeckinstance.com --backend_bearer_token=123
```

## Building
There is a Makefile in the project root, all you have to do is run `make all` which will build the mod and proxy server. Note that the device mod relies on Podman to build inside of a container environment (for convenience), but this can be changed to Docker if you prefer.

To build the container version of the proxy:

```sh
cd proxy-server
podman build .
```

## Acknowledgements
This project would not have been possible without all the hard work done by the `pgaskin` in creating [NickelHook](https://github.com/pgaskin/NickelHook), which is a wonderful framework for creating Kobo mods, and includes handy features like logging and failsafes.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

## Disclaimer

I work for Google, but this project is not an official Google project. It is not supported by Google and Google specifically disclaims all warranties as to its quality, merchantability, or fitness for a particular purpose.