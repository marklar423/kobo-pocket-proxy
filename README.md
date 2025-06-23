# Kobo Pocket Proxy

In light of the [Pocket shutdown](https://support.mozilla.org/en-US/kb/future-of-pocket), this project aims to preserve the wonderful Pocket integration on Kobo eReaders. 

This project has two main components:

## 1 - Kobo Mod
This is a modification to be installed on your physical Kobo eReader. It replaces all HTTP API calls to `getpocket.com` or `text.getpocket.com` to instead point to the URL of your choice, as specified by the config file, located at `/.adds/pocket_proxy/pocket_proxy.conf` on your device.

### Installation & Uninstallation
To install copy the KoboRoot.tgz file to the `/.kobo/` directory on your Kobo ereader, and unplug your device. It will install the modification and reboot.

To remove the mod, simply delete the file located at `/.adds/pocket_proxy/DELETE_ME_TO_UNINSTALL` on your device, and reboot the device.

## 2 - Proxy Server

If your chosen Pocket replacement already has a Pocket-compatible API, then simply edit the mod config file to point to the replacement's API. 

If not, then the included proxy server is designed to translate Pocket API calls to the API of one of the supported services, so the Kobo can keep talking to the new backend even after the Pocket API officially shuts down.

Currently, only Readeck is supported, but please feel free to contribute code for other backends.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

## Disclaimer

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.