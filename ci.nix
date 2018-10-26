{ pkgs ? import ./nix/nixpkgs.nix }: rec {
  meta = {
    name = "scylla";
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    docker-containers = [ "docker" ];
  };

  all = rec {
    scylla = pkgs.callPackage ./. { scylla-frontend = frontend; };
    docker = pkgs.callPackage ./nix/docker.nix { scylla = scylla.scylla; };
    depTree = scylla.depTree;
    hello = pkgs.hello;
    frontend = pkgs.callPackage ./frontend {};
    slowFailing = pkgs.runCommand "slow-failing" {} ''
      for i in {0..10..1}; do
        echo $i
        sleep 1
      done
    '';
    slowPassing = pkgs.runCommand "slow-passing" {} ''
      for i in {0..10..1}; do
        echo $i
        sleep 1
      done
      touch $out
    '';
  };

  scylla = all.scylla.scylla;
  hello = all.hello;
  docker = all.docker;
  deep = pkgs.recurseIntoAttrs { };
}
