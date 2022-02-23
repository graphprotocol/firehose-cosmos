# Cosmoshub4 Network

To start firehose for the Gaia/Cosmoshub4 network, first run the bootstraping script:

```bash
./bootstrap.sh
```

In case if you'd like to reset your local copy, remove the `./tmp` directory, or
run the script with an extra environment variable:

```bash
CLEANUP=1 ./bootstrap.sh
```

After bootstrapping is complete, start the firehose:

```bash
./start.sh
```

This will start the node from genesis, so give it some time until it start syncing.

You may check on the node's status (if its running) by opening `http://localhost:26657/status` in your browser.

Test if firehose is ready to stream blocks with `grpcurl` command:

```bash
grpcurl -plaintext localhost:9030 sf.firehose.v1.Stream.Blocks | jq
```

Make sure you have both `grpcurl` and `jq` commands installed.
