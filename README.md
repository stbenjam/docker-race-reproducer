# Reproducer for race in archive.DecompressStream

There is a race condition in pkg/archive when using `cmd.Start` for pigz
and xz. The `*bufio.Reader` could be returned to the pool while the
command is still writing to it. The command is wrapped in a
`CommandContext` where the process will be killed when the context is
cancelled, however this is not instantaneous, so there's a brief window
while the command could still be running but the `*bufio.Reader` was
already returned to the pool.

wrapReadCloser calls `cancel()`, and then `readBuf.Close()` which
eventually returns the buffer to the pool:

https://github.com/moby/moby/blob/1d19062b640b66daaf3e6f2246a947aaaf909dec/pkg/archive/archive.go#L179-L180

However, because cmdStream runs `cmd.Wait` in a go routine that we never
wait for to finish, it is not safe to return the reader to the pool yet.
We need to ensure we wait for `cmd.Wait` to finish!

## Run the reproducer

The reproducer attempts to decompress a 5mb gzipped file 10 times.

```
$ go run main.go
Waiting...
Done
Done
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x4b20b0]

goroutine 34 [running]:
bufio.(*Reader).fill(0xc0000fe300)
	/usr/lib/golang/src/bufio/bufio.go:100 +0xe0
bufio.(*Reader).WriteTo(0xc0000fe300, 0x554bc0, 0xc000118068, 0x7fb786381fb0, 0xc0000fe300, 0x4eb001)
	/usr/lib/golang/src/bufio/bufio.go:486 +0x259
io.copyBuffer(0x554bc0, 0xc000118068, 0x554a40, 0xc0000fe300, 0x0, 0x0, 0x0, 0x0, 0x0, 0xc00010c060)
	/usr/lib/golang/src/io/io.go:384 +0x352
io.Copy(0x554bc0, 0xc000118068, 0x554a40, 0xc0000fe300, 0x0, 0xc00003c7b8, 0x4ebb89)
	/usr/lib/golang/src/io/io.go:364 +0x5a
os/exec.(*Cmd).stdin.func1(0x0, 0x0)
	/usr/lib/golang/src/os/exec/exec.go:234 +0x55
os/exec.(*Cmd).Start.func1(0xc000110160, 0xc00011e120)
	/usr/lib/golang/src/os/exec/exec.go:400 +0x27
created by os/exec.(*Cmd).Start
	/usr/lib/golang/src/os/exec/exec.go:399 +0x5af
exit status 2
```

## Run with patch applied

```
$ git apply fix.patch
$ go run main.go
Waiting...
Done
Done
Done
Done
Done
Done
Done
Done
Done
Done
Complete
```
