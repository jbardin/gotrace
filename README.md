## gotrace
### tracing for go programs

gotrace annotates function calls in go source files with log statements on entry and exit.

	gotrace
	-exported
			only annotate exported functions
	-package
			show package name prefix on function calls
	-prefix string
			log prefix (default "\t")
	-returns
			show function return
	-w	re-write files in place

TODO:
- name function literals
- output filters
