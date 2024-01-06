## Podman Desktop Lima Examples

Run `limactl start template://podman` to create a Lima instance named "podman".

To open a shell, run `limactl shell podman` or `LIMA_INSTANCE=podman lima`.

* <https://podman-desktop.io/docs/lima/creating-a-lima-instance>

* <https://podman-desktop.io/docs/lima/creating-a-kubernetes-instance>

Empty OS templates:

- [ubuntu](../ubuntu.yaml) (same as default)

- [fedora](../fedora.yaml) (base for podman)

- [centos-stream](../centos-stream-9.yaml)

- [almalinux](../almalinux-9.yaml)

Example with running both Podman and Kubernetes (CRI-O):
- [crio](./crio.yaml)

Example with running both Docker Engine and Podman Engine:
- [both](./both.yaml)

Example with running AlmaLinux (el9) with "cockpit-podman":
- [alma](./alma.yaml)
