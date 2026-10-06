# KubeBreach

<p align="center">
  <img src=".github/logo.png" alt="KubeBreach logo" width="200">
</p>

KubeBreach is a command-line tool, written in Go, for probing exposed
Kubernetes components over the network. It scans a target for well-known
Kubernetes-related ports and can attempt two specific anonymous-access
techniques: command execution through an open Kubelet API, and reading keys
from an unauthenticated etcd instance.

> **Origin:** KubeBreach is a fork of
> [Kubestroyer](https://github.com/Rolix44/Kubestroyer) by Rolix, adapted for
> Kubernetes pentest training. The Go module path is `github.com/Auditeur/KubeBreach`
> and the program banner prints **KubeBreach v0.3.0**. The command's source
> directory is still `cmd/kubestroyer/` (directory rename pending).

## Features

All features below are implemented in the source tree:

- **Port scanning** of a predefined list of Kubernetes-related ports over
  HTTP/HTTPS (`pkg/scanning.go`, `utils/ports.go`).
- **Optional NodePort TCP scan** of the `30000–32767` range, enabled with
  `--node-scan`.
- **Anonymous Kubelet RCE** against the Kubelet API on port `10250`: lists
  running pods via `/runningpods/`, lets you interactively select a container,
  and runs a command through the `/run/` endpoint (`--anon-rce`).
- **Anonymous etcd read** against port `2379`: lists keys and prints the values
  you select (`--etcd`).
- **Multiple targets** supplied inline (comma-separated) or from a file (one
  target per line).
- **Interactive selection** of pods/containers and etcd keys via a terminal
  multi-select prompt.

## Requirements

- **Go 1.19** or later (declared in `go.mod`) to build from source.
- Network reachability to the target(s) on the relevant ports.
- A terminal, because the `--anon-rce` and `--etcd` modes use interactive
  multi-select prompts.

Direct dependencies (from `go.mod`):

| Module | Purpose in the code |
| --- | --- |
| `github.com/pborman/getopt/v2` | CLI flag parsing |
| `github.com/AlecAivazis/survey/v2` | Interactive multi-select prompts |
| `go.etcd.io/etcd/client/v3` | etcd v3 client (`--etcd`) |
| `go.uber.org/zap` | Logger passed to the etcd client |
| `google.golang.org/grpc` | etcd connection error classification |

## Installation

```bash
git clone <your-fork-url>
cd KubeBreach
go build -o kubebreach ./cmd/kubestroyer
```

The entry point is `cmd/kubestroyer/main.go` (the source directory still uses
the upstream name). There is no `Makefile`, release workflow, or prebuilt
binary in the repository, so building from source is the only documented
method.

## Usage

```
./kubebreach -t <target> [--node-scan] [--anon-rce [-x <command>]] [--etcd]
```

Flags are defined in `utils/config.go`:

| Flag | Short | Default | Description |
| --- | --- | --- | --- |
| `--target` | `-t` | *(none, required)* | Target as an IP, domain, comma-separated list, or a path to a file containing one target per line. |
| `--node-scan` | | `false` | Also run a TCP scan of the NodePort range `30000–32767`. Only applies to the default port-scan mode. |
| `--anon-rce` | | `false` | Attempt command execution through an open Kubelet API (port `10250`). |
| `-x` | `-x` | `cat /var/run/secrets/kubernetes.io/serviceaccount/token` | Command executed by `--anon-rce`. Has no effect unless `--anon-rce` is set. |
| `--etcd` | | `false` | Attempt an anonymous read of etcd (port `2379`). |

`-t` / `--target` is mandatory; invoking the tool without it causes `getopt`
to print usage and exit.

### Target resolution

The value of `-t` is first checked against the filesystem (`os.Stat`):

- If it is an existing path, it is read as a file and each line becomes a
  target.
- Otherwise it is split on commas and each element becomes a target.

Every target is then processed in turn.

### Mode selection

The three actions are selected as follows (`pkg/cmd.go`):

- **Default** (neither `--anon-rce` nor `--etcd`): the port scan runs.
- **`--anon-rce`**: the Kubelet RCE attempt runs; the port scan is **not** run.
- **`--etcd`**: the etcd read attempt runs; the port scan is **not** run.
- If both `--anon-rce` and `--etcd` are passed, both run and the port scan is
  skipped.
- `--node-scan` only affects the default port-scan mode.

## How it works

1. **Input parsing** – `utils.Config()` parses the flags; `pkg.Execute()`
   resolves `-t` into a list of targets (inline list or file).
2. **Port scan (default mode)** – For each target, Kubestroyer issues an
   HTTP `GET` to every port in its built-in known-ports map. Ports `443` and
   `10250` are probed over HTTPS (with certificate verification disabled);
   `10250` is probed on `/metrics`, all other ports on `/`. A port is reported
   open when the request returns without error. With `--node-scan`, a raw TCP
   dial is additionally attempted on each port in `30000–32767`.
3. **Kubelet RCE (`--anon-rce`)** – Kubestroyer requests
   `https://<target>:10250/runningpods/`, parses the returned pod list, and
   presents the container names for interactive selection. For each selected
   container it sends a `POST` to
   `https://<target>:10250/run/<namespace>/<pod>/<container>` with the body
   `cmd=<command>` and prints the response. A `Unauthorized` or `Forbidden`
   response body stops the program.
4. **etcd read (`--etcd`)** – Kubestroyer connects to `<target>:2379` with the
   etcd v3 client, lists all keys under `/` (keys-only), presents them for
   interactive selection, and prints the value of each selected key.
5. **Output** – Results and status messages are printed to standard output
   with ANSI colors; sections are separated by a divider line.

## Kubernetes

The tool interacts with the following Kubernetes-related network surfaces.

### Ports probed (default scan)

The known-ports map is defined in `utils/ports.go`:

| Port | Label in the code |
| --- | --- |
| 443 | Kubernetes API port |
| 2379 | Etcd |
| 6666 | Etcd |
| 4194 | cAdvisor for containers metrics |
| 6443 | Kubernetes API port |
| 8443 | Minikube API port |
| 8080 | Insecure API port |
| 10250 | Kubelet API anonymous port |
| 10255 | Kubelet API read only |
| 10256 | Kube proxy health check server |
| 9099 | Calico health check server |
| 6782 / 6783 / 6784 | Weave metrics/endpoints |
| 44134 | Tiller service listening |
| 30000–32767 | Possible NodePort service (added at startup) |

Because the `30000–32767` range is added to this map at startup, those ports
are also probed over HTTP during the default scan, independently of
`--node-scan`.

### Kubelet API (`--anon-rce`)

- `GET /runningpods/` on port `10250` to enumerate running pods and their
  containers.
- `POST /run/<namespace>/<pod>/<container>` on port `10250` with a
  `cmd=<command>` form body to execute a command inside a container.

This path works only when the Kubelet is configured to allow anonymous
authentication and to authorize the corresponding actions. Properly configured
clusters return `Unauthorized`/`Forbidden`, which the tool reports and then
exits.

### etcd (`--etcd`)

- Connects to port `2379` and performs a prefix, keys-only `GET` on `/`, then
  reads selected keys.

This works only against an etcd endpoint that accepts unauthenticated client
connections. etcd serving the Kubernetes API typically stores sensitive
cluster data, including Secrets.

## Security considerations

Kubestroyer is an offensive security / audit tool. Its two exploitation modes
actively interact with a target rather than passively observing it:

- `--anon-rce` **executes a command inside a container** on the target when the
  Kubelet allows anonymous access. The default command reads the pod's
  service-account token
  (`/var/run/secrets/kubernetes.io/serviceaccount/token`), which is sensitive
  material.
- `--etcd` can **read arbitrary keys** from an unauthenticated etcd, which may
  include cluster Secrets and other confidential data.
- All HTTPS requests are made with TLS certificate verification **disabled**
  (`InsecureSkipVerify: true`), so connections are not authenticated against
  the server certificate.
- The port scan and NodePort scan generate connection attempts to many ports
  on the target, which is detectable network activity.

The behaviors the tool relies on (anonymous Kubelet access, unauthenticated
etcd, exposed services) are **misconfigurations or intentional exposures of
Kubernetes components**, not inherent vulnerabilities of Kubernetes itself.

## Examples

Port scan of a single host (default mode):

```bash
./kubebreach -t 10.0.0.10
```

Port scan including the NodePort range:

```bash
./kubebreach -t cluster.example.com --node-scan
```

Scan several targets at once, or from a file:

```bash
./kubebreach -t 10.0.0.10,10.0.0.11
./kubebreach -t targets.txt
```

Attempt anonymous Kubelet RCE with the default command (reads the
service-account token), then with a custom command:

```bash
./kubebreach -t 10.0.0.10 --anon-rce
./kubebreach -t 10.0.0.10 --anon-rce -x "id"
```

Attempt an anonymous etcd read:

```bash
./kubebreach -t 10.0.0.10 --etcd
```

> Output depends entirely on the target environment. The `--anon-rce` and
> `--etcd` modes open an interactive multi-select prompt before acting on the
> results.

## Development

- Source layout:
  - `cmd/kubestroyer/main.go` – entry point and banner.
  - `pkg/cmd.go` – target resolution and mode dispatch.
  - `pkg/scanning.go` – port and NodePort scanning.
  - `pkg/expoit.go` – Kubelet RCE and etcd read.
  - `utils/` – flag configuration, known ports, banner strings.
- Build: `go build -o kubebreach ./cmd/kubestroyer`

## Testing

There are no test files (`*_test.go`) and no CI workflow in the repository.

## Limitations

The following limitations are visible in the code:

- The NodePort TCP scan uses `net.Dialer{Timeout: 5}`, i.e. a 5-nanosecond
  dial timeout, which is extremely short and may cause connections to fail
  before completing.
- The default port scan always probes the `30000–32767` range over HTTP (the
  range is part of the known-ports map), regardless of `--node-scan`.
- A port is considered "open" whenever the HTTP request returns without a
  transport error; this reflects reachability rather than a verified service
  identity.
- Several error conditions (e.g. a failed Kubelet request, an `Unauthorized`
  response, a file-open error) call `log.Fatal`, which terminates the program.
- The tool has no configurable timeout, output format, or logging options
  beyond the flags listed above, and no environment-variable or config-file
  support.
- Target resolution treats any value matching an existing filesystem path as a
  file, so a target string that happens to collide with a local path is read as
  a file.

## License

MIT License — Copyright (c) 2023 Rolix. See [LICENSE.txt](LICENSE.txt).

## Disclaimer

This tool is intended for authorized security testing and auditing only. Using
it against systems you do not own or do not have explicit permission to test may
be illegal. The authors and contributors accept no liability for misuse or for
any damage caused by this software. Use it only on infrastructure for which you
have proper authorization.
