# Funky - Function Proxy Server

A simple proxy server written in Go used to forward function invocations to language specific servers. Funky handles capturing stdout and stderr logs, function invocation timeouts and a limited amount of parallel function invocations.

Funky requires two environment variables:
  * SERVERS - a number of language specific servers to initalize to handle function invocations
  * SERVER_CMD - the command to run to start a function server e.g. `python3 main.py hello.handle`

Any request to the function server will try to invoke the function on any free server. The request will block if no server if idle and able to process the request.

## Development

Doing local development is easy.

1. Check out/clone the python3-base-image:
    ```
    cd ..
    git clone git@github.com:dispatchframework/python3-base-image.git
    cd -
    ```
2. Run funky locally (pointing to the example function):
    ```
    PORT=8080 TIMEOUT=30 SERVERS=1 SERVER_CMD="python3 ../python3-base-image/function-server/main.py example.handler.dummy" go run main.go
    ```

TODO: Funky needs logging!