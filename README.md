## gotrace
### tracing for go programs

gotrace annotates function calls in go source files with log statements on entry and exit.

    Usage of gotrace:
      -exits: show function exits
      -exported: only annotate exported functions


TODO:
- rewrite multiple files at a time
- name function literals
- automatically fix missing "log" imports
- create a custom logger, with configuration options
