{ pkgs ? import ./nix/nixpkgs.nix }: {
  meta = {
    name = "scylla";
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    docker-containers = [ "docker" ];
  };
  hello = pkgs.hello;
  slowFailing = pkgs.runCommand "slow-failing" {} ''
    for i in {0..10..1}; do
      echo $i
      sleep 1
    done
  '';
  ignored = {
    scylla = pkgs.callPackage ./. {};
    docker = pkgs.callPackage ./nix/docker.nix {};
  };
  # deep = pkgs.recurseIntoAttrs { };
}
