# Coreum Network

To start Firehose for the Coreum network, first run the bootstraping script:

```bash
./bootstrap.sh
```

If there are no built binaries four your OS, build them manually using
instructions: https://github.com/CoreumFoundation/coreum/
and move to the `tmp/cored`.

In case if you'd like to reset your local copy, remove the `./tmp` directory, or
run the script with an extra environment variable:

```bash
CLEANUP=1 ./bootstrap.sh
```

After bootstrapping is complete, start the Firehose:

```bash
./start.sh
```

This will start the node from genesis, so give it some time until it start syncing.

You may check on the node's status (if its running) by opening `http://localhost:26657/status` in your browser.

Test if Firehose is ready to stream blocks with `grpcurl` command (the syncing might take some time):

```bash
grpcurl -plaintext localhost:9030 sf.firehose.v2.Stream/Blocks
```

Make sure you have the `grpcurl` installed.
