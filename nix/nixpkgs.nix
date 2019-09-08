import (fetchTarball {
  url =
    "https://github.com/nixos/nixpkgs-channels/archive/54f385241e6649128ba963c10314942d73245479.tar.gz";
  sha256 = "0bd4v8v4xcdbaiaa59yqprnc6dkb9jv12mb0h5xz7b51687ygh9l";
}) {
  config = { };
  overlays = [ (import ./overlay.nix) ];
}
