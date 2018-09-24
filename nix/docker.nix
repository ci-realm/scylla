{ callPackage
, dockerTools
, busybox
, coreutils
, curl
, git-info
}:

let
  Labels = {
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    "com.xing.docker_build.target" = "prod";
    "com.xing.git.sha1"   = git-info "git rev-parse --verify HEAD" ./..;
    "com.xing.git.time"   = git-info "git show -s --format=%cI HEAD" ./..;
    "com.xing.git.remote" = git-info "git config --get remote.origin.url" ./..;
  };
  scylla = callPackage ./.. {};
  baseImage = dockerTools.buildImage {
    name = "quay.dc.xing.com/e-recruiting-api-team/scylla";
    config = {
      Env = [
        "PATH=${busybox}/bin:${curl}/bin:${coreutils}/bin"
      ];
    };
  };
in dockerTools.buildImage {
  fromImage = baseImage;
  name = "quay.dc.xing.com/e-recruiting-api-team/scylla";
  tag = Labels."com.xing.git.sha1";
  created = Labels."com.xing.git.time";
  config = {
    inherit Labels;
    EntryPoint = ["${scylla}/bin/scylla"];
    Env = [
      "PATH=${scylla}/bin:${busybox}/bin:${curl}/bin:${coreutils}/bin"
      "HOST=0.0.0.0"
      "PORT=80"
    ];
    ExposedPorts = {
      "80/tcp" = {};
    };
  };
}
