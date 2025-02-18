---
title: "Module log-shipper: advanced usage"
---

## How to enable debugging logs?

Set the debug option of the module to `true` in the Deckhouse configmap as the following snippet shows:
```yaml
logShipper: |
  debug: true
```

Then in logs, you will find a lot of helpful information about HTTP requests, connects reusing, detailed traces, and so on.

## How to get aware of logs pipelines?

To begin with, go to the command shell of the Pod on a desired node.
```bash
kubectl -n d8-log-shipper get pods -o wide | grep $node
kubectl -n d8-log-shipper exec $pod -it -c vector -- bash
```

All following commands are assumed to be executed from the Pod's shell.

### See pipelines as a graph

* Execute the `vector graph` command to get the output of logs pipelines topology in the [DOT format](https://graphviz.org/doc/info/lang.html).
* Put the output to [webgraphviz](http://www.webgraphviz.com/) os similar service to render the graph. 

Example of the graph output for a single pipeline in ASCII format:
```
+------------------------------------------------+
|  d8_cluster_source_flant-integration-d8-logs   |
+------------------------------------------------+
  |
  |
  v
+------------------------------------------------+
|       d8_tf_flant-integration-d8-logs_0        |
+------------------------------------------------+
  |
  |
  v
+------------------------------------------------+
|       d8_tf_flant-integration-d8-logs_1        |
+------------------------------------------------+
  |
  |
  v
+------------------------------------------------+
| d8_cluster_sink_flant-integration-loki-storage |
+------------------------------------------------+
```

### Investigate data processing

There is the `vector top` command to help you see how much data is going through all checkpoints of the pipeline.

Example of the output:

![Vector TOP output](../../images/460-log-shipper/vector_top.png)

### Get raw log samples

You can execute the `vector tap` to get all raw samples for all logging configs.
The only argument to the command is the ID of the pipeline stage (glob patterns are allowed).

Logs before applying any transforms:
```bash
vector tap d8_cluster_source_*
```

Transformed logs:
```bash
vector tap d8_tf_*
```

You can then use the `vector vrl` interactive console to debug [VRL](https://vector.dev/docs/reference/vrl/) remap rules for messages.

Example of a program on VRL:
```
. = {"test1": "lynx", "test2": "fox"}
del(.test2)
.
```

## How to add a new source/sink support for log-shipper?

Vector in the log-shipper module has been built with the limited number of enabled [features](https://doc.rust-lang.org/cargo/reference/features.html) (to improve building speed and decrease the size of the final binary).

You can see a list of all supported features by executing the `vector list` command.

If supporting a new source/sink is required, you need to add the corresponding feature to the list of enabled features in the Dockerfile.
