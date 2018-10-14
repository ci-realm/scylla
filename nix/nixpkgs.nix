import (
  fetchTarball {
    url = https://github.com/nixos/nixpkgs/archive/35bccdecc1ced35a62e0996c46330eed5d55097c.tar.gz;
    sha256 = "11gmrihncn4f88r0ak7ic9xb5ka3gln6qxyw7v4jm2dgp2qlr7c2";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
