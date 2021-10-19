# Targzsize

A quick tool to compute the total unpacked size of a set of tar.gz archives.

    targzsize [-legal] [-no-progress] [-human] path [path...]

Targzsize iterates over the provides paths and computes the unpacked size of each file within the packages archives.
It then adds these totals together and outputs it to standard output.

By default, targzsize writes status messages to standard error.
Pass the '-no-progress' flag to prevent this.

By default the standard output will contain a single number, representing the total size in bytes.
To instead use human readable units, pass the '-human' flag.
This flag also applies to status messages.

The '-legal' flag can be used to print legal and licensing information.

## License

MIT LICENSE