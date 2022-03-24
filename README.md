# MicroK8s cluster agent

This repository contains the code of the cluster agent service used in [MicroK8s](https://github.com/canonical/microk8s.git).

## Development Environment

The microk8s cluster agent requires [Go](https://go.dev). Install Go with:

```bash
sudo snap install go --classic
```

Run all tests with:

```bash
make go.lint go.test go.vet go.staticcheck
```

## Build

```bash
make cluster-agent
./cluster-agent --help
```
