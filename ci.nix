{ pkgs ? import ./nix/nixpkgs.nix }: rec {
  meta = {
    name = "scylla";
    maintainer = "Michael Fellinger <scylla@manveru.dev>";
    docker-containers = [ "docker" ];
  };

  all = rec {
    scylla = pkgs.callPackage ./. { inherit frontend; };
    scyllaDB = pkgs.copyPathToStore ./db;
    docker = pkgs.callPackage ./nix/docker.nix { scylla = scylla.scylla; };
    depTree = scylla.depTree;
    hello = pkgs.hello;
    frontend = pkgs.callPackage ./frontend { pkgs = pkgs; };
    slowFailing = pkgs.runCommand "slow-failing" { } ''
      for i in {0..60..1}; do
        echo $i
        sleep 1
      done
    '';
    slowPassing = pkgs.runCommand "slow-passing" { } ''
      for i in {0..60..1}; do
        echo $i
        sleep 1
      done
      touch $out
    '';
  };

  scylla = all.scylla.scylla;
  scyllaDB = all.scyllaDB;
  frontend = all.frontend;
  hello = all.hello;
  docker = all.docker;
  deep = pkgs.recurseIntoAttrs { };
}
