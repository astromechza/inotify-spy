# inotify-spy

Binary for watching inotify events on a folder recursively. Similar to inotifywatch.

## Usage

```bash
$ ./inotify-spy --help
inotify-spy is a simple binary for doing inotify watching on Linux.
It will allow you to watch a directory or directory tree for file events:

- Create
- Write
- Remove
- Rename
- Chmod
- Open

Because this tool uses inotify events, it has to create one for each directory
that it watches. On most systems there is a limit to the number of inotify
watches or files a process is allowed to create at once. You might hit these
limits if you try to watch a very large tree of directories. On most systems
there are ways to increase these limits if required.

This tool will continue to watch the tree until you stop it using SIGINT (Ctrl-C)
which will cause it to print a sorted summary of the files touched.

Usage: inotify-spy [-live] [-mute-errors] [-recursive] directory

  -dont-record-chmod
        Don't record chmod events
  -dont-record-create
        Don't record create events
  -dont-record-open
        Don't record open events
  -dont-record-remove
        Don't record remove events
  -dont-record-rename
        Don't record rename events
  -dont-record-write
        Don't record write events
  -export-csv string
        Export summary as csv to the given path
  -ignore-prefixes string
        File to read ignore prefixes from
  -live
        Show events live, not just as a summary at the end
  -mute-errors
        Mute error messages related to setting up watches
  -recursive
        Recursively watch target directory
  -sort-name
        Sort summary by file path rather than most events
  -version
        Print version information
```

## Example

Lets create a working directory:

```
$ mkdir -p testing/childdir/grandchilddir
$ cd testing
```

Now we start up inotify-spy and press Enter to start recording:

```
$ inotify-spy -recursive .
Beginning to watch events..
Watching 3 directories..
Press enter to start recording:

```

And do some operations:

```
touch bob
echo "something" > bob
cp bob john
mv bob childdir/bob
cat childdir/bob
echo "another thing" > childdir/grandchilddir/charles
rm childdir/bob
```

And then stop inotify-spy and see what we get:

```
Beginning to watch events..
Watching 3 directories..
Press enter to start recording:

Beginning to record events. Press Ctrl-C to stop..
^CReceived interrupt signal. Stopping.

Create Write  Remove Rename Chmod  Open   Path
1      2      0      1      1      3      /home/bmeier/testing/bob
1      1      0      0      0      1      /home/bmeier/testing/childdir/grandchilddir/charles
1      1      0      0      0      1      /home/bmeier/testing/john
1      0      1      0      0      1      /home/bmeier/testing/childdir/bob
```

If we ran it with `-live` we would also see:

```
Beginning to record events. Press Ctrl-C to stop..
event: "/home/bmeier/testing/bob": CREATE
event: "/home/bmeier/testing/bob": OPEN
event: "/home/bmeier/testing/bob": CHMOD
event: "/home/bmeier/testing/bob": WRITE
event: "/home/bmeier/testing/bob": OPEN
event: "/home/bmeier/testing/bob": WRITE
event: "/home/bmeier/testing/bob": OPEN
event: "/home/bmeier/testing/john": CREATE
event: "/home/bmeier/testing/john": OPEN
event: "/home/bmeier/testing/john": WRITE
event: "/home/bmeier/testing/bob": RENAME
event: "/home/bmeier/testing/childdir/bob": CREATE
event: "/home/bmeier/testing/childdir/bob": OPEN
event: "/home/bmeier/testing/childdir/grandchilddir/charles": CREATE
event: "/home/bmeier/testing/childdir/grandchilddir/charles": OPEN
event: "/home/bmeier/testing/childdir/grandchilddir/charles": WRITE
event: "/home/bmeier/testing/childdir/bob": REMOVE
^CReceived interrupt signal. Stopping.
```

### Using ignore prefixes

In some cases, mostly very large and deep directory trees, or systems with
alot of other noisey file changes, you may want to ignore events from some
parts of the tree.

To do this, `inotify-spy` supports an `-ignore-prefixes` option. This option
should point to a file that contains a list of absolute path prefixes that will
be ignored. I chose to use absolute path prefixes in this file, since relative
paths lose their meaning when calling the binary from different locations or
when changing directories.

For example, we want to watch `/var/` but we are getting too many unwanted
events from an application logging to `/var/log`. Ignore prefixes allow us
to create a file containing `/var/log` and specify that at runtime:

```
$ inotify-spy -recursive -live -ignore-prefixes prefixe-file /var
```

When a path is being ignored, you'll see something like this:

```
Not watching /var/log or its children since it matches an ignore prefix
```
