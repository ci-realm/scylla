{ callPackage
, lib
, stdenv
, dockerTools
, writeTextFile
, busybox
, coreutils
, curl
, git-info
, cacert
, git
, gnutar
, which
, openssh
, vim
, bashInteractive
, nixUnstable
}:

let
  # TODO: this isn't in nixpkgs proper yet
  graham = import (fetchTarball {
    # url = https://github.com/graham-at-target/nixpkgs/archive/0e23958981e3a99567dbcbdadf694384e8022a47.tar.gz;
    url = https://github.com/NixOS/nixpkgs/archive/4fe900619080f0b1a804abbccaab500d819ead10.tar.gz;
  }) {};

  inherit (graham.dockerTools) buildLayeredImage buildImage;

  scylla = (callPackage ./.. {}).scylla;

  executables = [
    bashInteractive
    busybox
    coreutils
    curl
    git
    gnutar
    nixUnstable
    openssh
    which
    vim
  ];

  labels = {
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    "com.xing.docker_build.target" = "prod";
    "com.xing.git.sha1"   = git-info "git rev-parse --verify HEAD" ./..;
    "com.xing.git.time"   = git-info "git show -s --format=%cI HEAD" ./..;
    "com.xing.git.remote" = git-info "git config --get remote.origin.url" ./..;
  };

in buildLayeredImage {
  name = "quay.dc.xing.com/e-recruiting-api-team/scylla";
  tag = git-info "git rev-parse --verify HEAD" ./..;
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
