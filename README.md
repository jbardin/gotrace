## gotrace
### tracing for go programs

gotrace annotates function calls in go source files with log statements on entry and exit.

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
		log prefix (default "\t")
	-returns
		show function return
	-timing
		print function durations. Implies -returns
	-w	re-write files in place

