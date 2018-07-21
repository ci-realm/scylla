{ callPackage, dockerTools, git-info }:
let
  Labels = {
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    "com.xing.docker_build.target" = "prod";
    "com.xing.git.sha1"   = git-info "git rev-parse --verify HEAD" ./..;
    "com.xing.git.time"   = git-info "git show -s --format=%cI HEAD" ./..;
    "com.xing.git.remote" = git-info "git config --get remote.origin.url" ./..;
  };
  scylla = callPackage ./.. {};
in
dockerTools.buildImage {
  name = "scylla";
  tag = "production";
  created = Labels."com.xing.git.time";
  config = {
    inherit Labels;
    WorkingDir = "/";
    EntryPoint = ["${scylla}/bin/scylla"];
    Env = [
      "HOST=0.0.0.0"
      "PORT=80"
    ];
    ExposedPorts = {
      "80/tcp" = {};
    };
  };
  contents = [ scylla ];
}
