## gotrace
### tracing for go programs

gotrace annotates function calls in go source files with log statements on entry and exit.

    Usage of gotrace:
    -exits
    		show function exits
    -exported
    		only annotate exported functions
    -package
    		show package name prefix on function calls
    -prefix string
    		log prefix (default "\t")

TODO:
- rewrite multiple files at a time
- name function literals
- output filters
