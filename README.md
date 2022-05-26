# MicroK8s cluster agent

This repository contains the code of the cluster agent service used in [MicroK8s](https://github.com/canonical/microk8s.git).

## Development Environment

The microk8s cluster agent requires [Go](https://go.dev). Install Go with:

```bash
sudo snap install go --classic
```

Run all tests with:

```bash
make go.fmt go.lint go.test go.vet go.staticcheck
```

## Build

```bash
make cluster-agent
./cluster-agent --help
```

## Testing with live MicroK8s instance

### MicroK8s running on local machine

```bash
cd cluster-agent
sudo snap download microk8s --basename=microk8s
sudo unsquashfs ./microk8s.snap
sudo chown -R 1000: squashfs-root
sudo sed 's,$SNAP/bin/cluster-agent,go run '"$PWD"',' -i squashfs-root/run-cluster-agent-with-args

# development iteration
sudo snap restart microk8s.daemon-cluster-agent
```

### MicroK8s running on remote machine

```bash
export HOST=ubuntu@10.0.0.100

ssh $HOST '
    sudo snap download microk8s --basename=microk8s
    sudo unsquashfs microk8s.snap
    sudo chown -R 1000: squashfs-root
    sudo snap try ./squashfs-root
'

# development iteration
go build .
ssh $HOST 'sudo snap stop microk8s.daemon-cluster-agent'
scp ./microk8s-cluster-agent $HOST:squashfs-root/bin/cluster-agent
ssh $HOST 'sudo snap start microk8s.daemon-cluster-agent'
```
