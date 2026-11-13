#!/bin/bash

set -euo pipefail

REGISTRY_DIR="{{ .RegistryDir }}"
PAYLOAD="{{ .Payload }}"

for arg in $(</proc/cmdline); do
  case "${arg}" in
    {{ .CmdlineKey }}=*) PAYLOAD="${arg#{{ .CmdlineKey }}=}" ;;
  esac
done

if [[ -d "${REGISTRY_DIR}/registry" ]]; then
  echo "Registry payload already in place"
  exit 0
fi

mkdir -p "${REGISTRY_DIR}"
WORKDIR=$(mktemp -d "${REGISTRY_DIR}/.payload.XXXXXX")
trap 'rm -rf "${WORKDIR}"' EXIT

echo "Fetching registry payload from ${PAYLOAD}"

case "${PAYLOAD}" in
  http://*|https://*)
    curl --fail --silent --show-error --location --retry 10 --retry-delay 5 --retry-connrefused "${PAYLOAD}" \
      | tar -x -C "${WORKDIR}"
    ;;
  LABEL=*)
    DEVICE="/dev/disk/by-label/${PAYLOAD#LABEL=}"
    for _ in $(seq 60); do
      [[ -e "${DEVICE}" ]] && break
      sleep 2
    done
    MOUNTPOINT=$(mktemp -d)
    mount -o ro "${DEVICE}" "${MOUNTPOINT}"
    cp -a "${MOUNTPOINT}/." "${WORKDIR}/"
    umount "${MOUNTPOINT}"
    rmdir "${MOUNTPOINT}"
    ;;
  *)
    echo "Error: unsupported registry payload source '${PAYLOAD}'" >&2
    exit 1
    ;;
esac

if [[ ! -d "${WORKDIR}/registry" ]]; then
  echo "Error: registry payload from ${PAYLOAD} contains no registry directory" >&2
  exit 1
fi

mv "${WORKDIR}"/* "${REGISTRY_DIR}/"
