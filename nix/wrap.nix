{ lib, bubblewrap, busybox, bashInteractive, curl, nix, closureInfo
, writeShellScriptBin, writeReferencesToFile }:

let
  contents = [ bashInteractive busybox curl nix ];
in writeShellScriptBin "wrap.sh" ''
  #!/usr/bin/env bash

  set -euo pipefail

  rm -rf /tmp/scylla-container
  mkdir -p /tmp/scylla-container
  export NIX_REMOTE=local?root=/tmp/scylla-container
  ${nix}/bin/nix-store --load-db < ${
    closureInfo { rootPaths = contents; }
  }/registration

  for i in $(< ${closureInfo { rootPaths = contents; }}/store-paths); do
    cp -a "$i" /tmp/scylla-container/"''${i:1}"
  done

  chmod -R ug+w /tmp/scylla-container

  tree -L 3 /tmp/scylla-container

  (
    exec env --ignore-environment ${bubblewrap}/bin/bwrap \
      --ro-bind /usr /usr \
      --dir /tmp \
      --dir /var \
      --dir /build \
      --symlink ../tmp var/tmp \
      --proc /proc \
      --dev /dev \
      --ro-bind /etc/resolv.conf /etc/resolv.conf \
      --ro-bind /tmp/scylla-container/nix /nix \
      --ro-bind /usr/bin /usr/bin \
      --ro-bind /bin /bin \
      --chdir /build \
      --unshare-all \
      --share-net \
      --die-with-parent \
      --dir /run/user/$(id -u) \
      --setenv PATH ${lib.makeBinPath (contents ++ [ "" ])} \
      --setenv XDG_RUNTIME_DIR "/run/user/`id -u`" \
      --setenv PS1 "$ " \
      --file 11 /etc/passwd \
      --file 12 /etc/group \
      ${bashInteractive}/bin/sh
  ) \
  11< <(getent passwd $UID 65534) \
  12< <(getent group $(id -g) 65534)
''
