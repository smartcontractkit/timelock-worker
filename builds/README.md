The docker compose project in allow you to run the timelock worker either
by itself or with a prometheus instance to monitor the worker.

To run the worker by itself, copy the `.env.sample` file into a new `.env` file
in the `/builds` directory and specify the `ENV_FILE` variable. Then from the
`/builds` directory run the following command:

```bash
docker compose up -d
```

To run the worker with the prometheus instance, from the `/builds` directory run
the following command:

```bash
COMPOSE_PROFILES=metrics docker compose up -d
```

## Worker Changes

Currently any changes to the timelock worker requires that the container be rebuilt
manually. To do this, from the `/builds` directory run the following command:

```bash
docker compose build
```
