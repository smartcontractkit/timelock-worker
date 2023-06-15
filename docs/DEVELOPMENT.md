# Development documentation

The daemon is composed by an **event listener**, a **scheduler**, and an **executor**.

## **Event Listener**

It's constantly watching for [events](timelock.abi.md), and based on the events received it performs different actions:

- **CallScheduled**: the daemon store these inside the scheduler, to be executed later.
- **CallExecuted**: this event means that a CallSchedule event has been executed, hence the daemon has to remove this event from the scheduler.
- **Cancelled**: same as with CallExecuted, this event means the CallSchedule operation with that ID has been cancelled, hence the scheduler removes it.

## **Scheduler**

The scheduler maintains a store of **unexecuted CallSchedule** operations.

Every certain amount of seconds, defined by the poll period, it checks the operations in the store and sends them to the executor.

Also, whenever the timelock-worker goes down, either because of an error or manual action, it creates a file in `/tmp/timelock-worker.log`, containing all the pending CallScheduled transactions in the store, and the block number where those transactions reside.

As an example of `timelock-worker.log`:

```text
Process stopped at 2023-05-10 07:20:49.435871 +0000 UTC
Earliest CallSchedule pending ID: 14a3e53b7a9575d15dd7dec0cbc3bc71c0c58c099ae71c06f6d23c3ff23562f1      Block Number: 8935613
        Use this block number to ensure all pending operations are properly executed.
        Set it as environment variable or in timelock.env with FROM_BLOCK=8935613, or as a flag with --from-block=8935613
CallSchedule pending ID: 6444b766232d9bacba978f3a03eb7252860584e0d55b5ca2f29b105e2dceb3c7       Block Number: 8935880
CallSchedule pending ID: d9a1243050dc90e5926d7e9d88244ee0bd54a5bdb67b6235e958de92e84304ad       Block Number: 8936764
CallSchedule pending ID: d91015abae4d4645a7f0269c5455897df8a605097723657bdf812a05d3ad20de       Block Number: 8940349
```

Based on this data, the server can be instructed to start listening for events in the earliest CallSchedule event pending. In this example, that would be 8935613.

## **Executor**

The executor is where all the logic regarding operations exist.

It gets the pending operations sent by the scheduler, and checks the following:

- The operation is in Ready status, otherwise the operation is scheduled to the next cycle.
  - Note: The operation could remain pending either because the delay period hasn't passed yet, or the predecessor operation is still pending.

Based on that data, the operation is executed or sent back to the scheduler, to be checked in the next cycle.

[< back](README.md)
