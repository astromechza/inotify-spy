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

  -live
        Show events live, not just as a summary at the end
  -mute-errors
        Mute error messages related to setting up watches
  -recursive
        Recursively watch target directory
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

## Features to add

- ignore certain paths (some kind of ignore file)
- export as csv report
- force dontRecordOpen on all but linux
