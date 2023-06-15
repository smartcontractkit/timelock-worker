# Configuration

- [Configuration](#configuration)
  - [Summary of configuration options](#summary-of-configuration-options)
  - [Configuration file](#configuration-file)
  - [Environment variables](#environment-variables)
  - [Command flags](#command-flags)
    - [Available flags](#available-flags)

## Summary of configuration options

The `timelock-worker` requires certain values to be provided in order to run properly.

- **Node URL**: WS URL where the Timelock contract is deployed. A websocket connection is needed for the events listener.
- **Timelock address**: the target Timelock contract to watch.
- **Call proxy address**: address of the contract with EXECUTOR role in the Timelock contract.
- **Private Key**: this private key that can sign transactions in the Timelock contract.
- **From Block**: 0 by default, it's optional. The worker can be run to watch for event starting at an specified block.
- **Poll Period**: Scheduler period, in seconds, to go over all the pending operations.
- **Log Level**: info by default, and doesn't need to be changed for normal usage. It can be info or debug.

The configuration can be provided via configuration file `timelock.env`, `environment variables` or `flags`, or any combination of those.

The **precedence order** is `command flags > environment variables > configuration file`.

## Configuration file

The file is named `timelock.env`. Create the file in the **same directory** where the `timelock-worker` binary is, with the following format:

```text
# RPC WebSocket URL.
NODE_URL=wss://myrpchost/foo/bar

# Timelock contract address to monitor.
TIMELOCK_ADDRESS=0x12345

# Contract address with Executor role in TIMELOCK_ADDRESS.
CALL_PROXY_ADDRESS=0x67890

# EOA private key.
PRIVATE_KEY=MyPrivateKey

# Subscription starting block.
FROM_BLOCK=0

# Event polling period in seconds.
POLL_PERIOD=600
```

## Environment variables

It's possible also to set environment variables, which will take precedence over the configuration file, overriding the values.

The list of available environment variables is:

```text
NODE_URL
TIMELOCK_ADDRESS
CALL_PROXY_ADDRESS
PRIVATE_KEY
FROM_BLOCK
POLL_PERIOD
```

Example:

- Using the timelock.env file above, and setting the following environment var, would result in using the timelock address 0x09876 instead of 0x12345.

```bash
export TIMELOCK_ADDRESS=0x09876
```

- When running in a container, the environment variables can be set with `-e ENVVAR=VALUE`:

```bash
docker run -ti \
-e LOGLEVEL=info \
-e TIMELOCK_ADDRESS=TimelockAddress \
-d timelock-worker:latest
```

## Command flags

Lastly, the `timelock-worker` also can be provided configuration with flags, which take precedence over timelock.env and environment variables.

- When running `timelock-worker`, specify the flags as usual.

```bash
timelock-worker start -n wss://rpc-url -k PrivateKey -a 0x12345
```

- When running `timelock-worker` as a container image, the flags can be specified as well in the following format.

```bash
 docker run -d -it timelock-worker:latest \
 start --log-level=info --private-key MyPrivKey
```

Note that, if specifying a flag in docker mode, start has to be also in the command execution: `start --log-level=info --private-key MyPrivKey`

### Available flags

```bash
Usage:
  timelock-worker start [flags]

Flags:
  -f, --call-proxy-address string   Address of the target CallProxyAddress contract (default "0x67890")
      --from-block int              Start watching from this block
  -h, --help                        help for start
  -n, --node-url string             RPC Endpoint for the target blockchain (default "wss://myrpchost/foo/bar")
      --poll-period int             Poll period in seconds (default 600)
  -k, --private-key string          Private key used to execute transactions (default "MyPrivateKey")
  -a, --timelock-address string     Address of the target Timelock contract (default "0x12345")

Global Flags:
  -l, --log-level string   set logging level (debug, info) (default "info")
  -o, --output string      set logging output (human, json) (default "human")
```

[< back](README.md)
