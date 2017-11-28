# gotrace

## Tracing for go programs

**gotrace** annotates function calls in go source files with log statements on entry and exit.

```
usage: gotrace [flags] [path ...]
-exclude string
	exclude any matching functions, takes precedence over filter
-exported
	only annotate exported functions
-filter string
	only annotate functions matching the regular expression (default ".")
-formatLength int
	limit the formatted length of each argument to 'size' (default 1024)
-package
	show package name prefix on function calls
-prefix string
	log prefix
-returns
	show function return
-timing
	print function durations. Implies -returns
-w	re-write files in place
```

### Example

```sh
    # gotrace operates directly on go source files.
    # Insert gotrace logging statements into all *.go files in the current directory
	# Make sure all files are saved in version control, as this rewrites them in-place!

    $ gotrace -w -returns ./*.go
```
