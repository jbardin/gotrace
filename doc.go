/*

Command gotrace annotates function entry and exit points to provide strace-like
tracing of go programs.

	usage: gotrace [flags] [path ...]
	-exclude string
			exclude any matching functions, takes precedence over filter
	-exported
			only annotate exported functions
	-filter string
			only annotate functions matching the regular expression (default ".")
	-package
			show package name prefix on function calls
	-prefix string
			log prefix (default "\t")
	-returns
			show function return
	-w	re-write files in place

*/
package main
