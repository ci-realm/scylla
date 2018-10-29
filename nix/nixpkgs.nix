import (
  fetchTarball {
    url = https://github.com/NixOS/nixpkgs/archive/e8df5045cac40f9919e7b14d356308ace53a947e.tar.gz;
    sha256 = "09k3azns0cwhndgklz5yhkka7a509s5nyvnfz30pvrl1gj4qqg2w";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
