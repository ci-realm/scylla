import (
  fetchTarball {
    url = https://github.com/manveru/nixpkgs/archive/3a275ac1eb4ea538222e8b282d9804bab6e6e543.tar.gz;
    sha256 = "19xm135jxg9n8bkm5ncwr3bq41jp93nhw0jh81c0rh7aq0cahl4r";
  }
) {
  config = {};
  overlays = [
    (import ./overlay.nix)
  ];
}
