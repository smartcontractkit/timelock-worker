# timelock-worker builds

- [timelock-worker builds](#timelock-worker-builds)
  - [Build as a command line interface](#build-as-a-command-line-interface)
  - [Build as a container image](#build-as-a-container-image)

## Build as a command line interface

- Run the `build` target in the `Makefile`.

```bash
make build
```

- The executable will be placed in `bin/timelock-worker`

## Build as a container image

- An **example** `Dockerfile` is provided in `builds/Dockerfile`.
  - Note that it is just **an example** on how to build a container with timelock-worker, and it's **not supported**.

```bash
docker buildx build -f builds/Dockerfile -t timelock-worker:latest --load .
```

- The resulting image `timelock-worker:latest` will be saved into docker. Check it with:

```bash
docker images
```

- The output will show something similar to the following:

```bash
REPOSITORY        TAG               IMAGE ID       CREATED        SIZE
timelock-worker   latest            6b59e125d23f   26 hours ago   367MB
```

[< back](README.md)
