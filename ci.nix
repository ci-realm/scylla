{ pkgs ? import ./nix/nixpkgs.nix }: {
  meta = {
    name = "scylla";
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    docker-containers = [ "docker" ];
  };
  scylla = pkgs.callPackage ./. {};
  docker = pkgs.callPackage ./nix/docker.nix {};
  deep = pkgs.recurseIntoAttrs {
    scylla = pkgs.callPackage ./. {};
    thisIsNotEvaluated = {
      scylla = pkgs.callPackage ./. {};
    };
  };
}
