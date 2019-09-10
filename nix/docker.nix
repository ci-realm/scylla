{ callPackage
, lib
, stdenv
, writeTextFile
, dockerTools
, busybox
, coreutils
, curl
, cacert
, git
, gnutar
, which
, openssh
, vim
, bashInteractive
, nix
, scylla
}:

let
  inherit (dockerTools) buildLayeredImage buildImage;

  executables = [
    bashInteractive
    busybox
    coreutils
    curl
    git
    gnutar
    nix
    openssh
    which
    vim
  ];

  labels = {
    maintainer = "Michael Fellinger <scylla@manveru.dev>";
  };

in buildLayeredImage {
  name = "quay.dc.xing.com/e-recruiting-api-team/scylla";
  tag = "latest";
  created = "now";
  maxLayers = 90;
  contents = [ # FIXME: graham has a patch for this he'll push soon
    (writeTextFile { name = "passwd"; text = "root:x:0:0:root:/:/bin/sh"; destination = "/etc/passwd"; })
    (writeTextFile { name = "nix.conf"; text = "build-users-group ="; destination = "/etc/nix/nix.conf"; })
  ];
  config.Cmd = [ "${scylla}/bin/scylla" ];
  config.Labels = labels;
  config.ExposedPorts."80/tcp" = {};
  config.Env = [
    "SSL_CERT_FILE=${cacert}/etc/ssl/certs/ca-bundle.crt"
    "PATH=${lib.makeBinPath executables}"
    "HOST=0.0.0.0"
    "PORT=80"
    "HOME=/"
    "BUILD_DIR=/ci"
    "PREPARE_KNOWN_HOSTS=true"
  ];
}
