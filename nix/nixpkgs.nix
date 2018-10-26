import (
  fetchTarball {
    url = https://github.com/NixOS/nixpkgs/archive/8070a6333f3fc41ef93c2b0e07f999459615cc8d.tar.gz;
    sha256 = "0v6nycl7lzr1kdsy151j10ywhxvlb4dg82h55hpjs1dxjamms9i3";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
