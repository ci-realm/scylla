{ pkgs ? import ./nix/nixpkgs.nix }: {
  meta = {
    name = "scylla";
    maintainer = "Michael Fellinger <mf@seitenschmied.at>";
    docker-containers = [ "docker" ];
  };
  hello = pkgs.hello;
  ignored = {
    scylla = pkgs.callPackage ./. {};
    docker = pkgs.callPackage ./nix/docker.nix {};
  };
  # deep = pkgs.recurseIntoAttrs { };
}
