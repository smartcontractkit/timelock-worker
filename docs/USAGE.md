# Usage

## Recommended usage

### As a command line interface

- Run the `timelock-worker` with the combination of timelock.env, environment variables and/or flags wanted.
- There's no restriction on wether to use only one method or a combination of any number of them to run the server.

Example usage with command line:

```bash
# in this example, the administrator decides to not provide a timelock.env, but instead provide configuration via environment variables, except for poll-period, which will be provided via flag
$ export TIMELOCK_ADDRESS=0x776FC69c49AA3c067e92B4155394c3027D5B51b8
$ export CALL_PROXY_ADDRESS=0x776FC69c49AA3c067e92B4155394c3027D5B51b8
$ export PRIVATE_KEY=04491a4c29c64295b453e53831ab633686b26d8f36a8d0735b1f4bb867f8af78
$ export NODE_URL=wss://sepolia.infura.io/ws/v3/266402bb492e4276973d3c1b380680fa
$ export FROM_BLOCK=3510520

# start the timelock-worker
$ bin/timelock-worker start --log-level debug --poll-period 1
```

Remember that any combination of environment variables, flags and configuration in timelock.env can be provided.

The **precedence order** is `command flags > environment variables > configuration file`.

### As a container image

- Note that, when exiting, `timelock-worker` will leave logs in `/tmp/timelock-worker`. And when working inside a container, that filesystem is inside the container. So, to preserve the log, docker can share a volume with `-v`.
- Also, if the administrator wants to provide its own `timelock.env` configuration file, the container will expect a file in `/app/timelock.env`, which can be provided with `mount`.

Example using docker:

- Provide your own timelock.env and the current directory, so the container can leave the log after exiting.

```bash
 docker run -d -it \
 # the file in the current-directory/timelock.env will be mounted inside the container in /app/timelock.env
 --mount type=bind,source="$(pwd)"/timelock.env,target=/app/timelock.env,readonly \
 # when the process exits, the timelock-worker.log will be placed in the current working directory
 -v "$(pwd)":/tmp \
 timelock-worker:latest
```

- Providing some environment variables **and** timelock.env

```bash
 docker run -d -it \
 -e LOGLEVEL=info \
 -e TIMELOCK_ADDRESS=TimelockAddress \
 # the file in the current-directory/timelock.env will be mounted inside the container in /app/timelock.env
 --mount type=bind,source="$(pwd)"/timelock.env,target=/app/timelock.env,readonly \
 # when the process exits, the timelock-worker.log will be placed in the current working directory
 -v "$(pwd)":/tmp \
 timelock-worker:latest
```

- Providing environment variables, timelock.env, and flags.

```bash
 docker run -d -it \
 -e LOGLEVEL=info \
 -e TIMELOCK_ADDRESS=TimelockAddress \
 # the file in the current-directory/timelock.env will be mounted inside the container in /app/timelock.env
 --mount type=bind,source="$(pwd)"/timelock.env,target=/app/timelock.env,readonly \
 # when the process exits, the timelock-worker.log will be placed in the current working directory
 -v "$(pwd)":/tmp \
 # log level debug takes precedence over log level info in env var.
 timelock-worker:latest start -log-level=debug
```

[< back](README.md)
