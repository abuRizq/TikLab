#!/bin/bash
set -e

# Workdir is mounted at /data: base/ and overlay/ subdirs
DATA="/data"
OVERLAY="$DATA/overlay/disk.qcow2"
BASEDIR="$DATA/base"

if [ ! -f "$OVERLAY" ]; then
  echo "Fatal: overlay not found at $OVERLAY"
  exit 1
fi

# Find first .qcow2 or .img in base/
BASE=""
for f in "$BASEDIR"/*.qcow2 "$BASEDIR"/*.img; do
  [ -f "$f" ] && { BASE="$f"; break; }
done
if [ -z "$BASE" ]; then
  echo "Fatal: no base image (chr.qcow2 or chr.img) in workdir/base/"
  exit 1
fi

# QEMU: 1G RAM, 1 CPU, virtio disk, no display, serial console
# Port forwarding: API 8728, Winbox 8291, SSH 22 -> host 2222
KVM=""
[ -e /dev/kvm ] && KVM="-enable-kvm"
exec qemu-system-x86_64 \
  $KVM \
  -m 1024 \
  -smp 1 \
  -drive file="$OVERLAY",if=virtio,format=qcow2 \
  -nographic \
  -device virtio-net-pci,netdev=net0 \
  -netdev user,id=net0,hostfwd=tcp::8728-:8728,hostfwd=tcp::8291-:8291,hostfwd=tcp::2222-:22
