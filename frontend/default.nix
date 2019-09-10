{ pkgs ? import <nixpkgs> { } }:

let
  inherit (pkgs) stdenv elmPackages lib;

  keepPrefixes = (map (pa: toString pa) [ ./js ./src ./elm.json ./index.html ]);

  mkDerivation = { srcs ? ./elm-srcs.nix, src, name, srcdir ? "./src"
    , targets ? [ ], versionsDat ? ./versions.dat }:
    stdenv.mkDerivation {
      inherit name src;

      buildInputs = [ elmPackages.elm ];

      buildPhase = pkgs.elmPackages.fetchElmDeps {
        elmPackages = import srcs;
        inherit versionsDat;
      };

      installPhase = let
        elmfile = module:
          "${srcdir}/${builtins.replaceStrings [ "." ] [ "/" ] module}.elm";
      in ''
        mkdir -p $out/share/doc
        cp -r ${./static} $out/static
        chmod -R u+w $out

        ${lib.concatStrings (map (module: ''
          echo "compiling ${elmfile module}"
          elm make ${elmfile module} \
            --optimize \
            --output $out/static/js/${module}.js \
            --docs $out/share/doc/${module}.json
        '') targets)}
      '';
    };
in mkDerivation {
  name = "scylla-frontend-0.1.0";
  srcs = ./elm-srcs.nix;
  src = __filterSource
    (path: type: __any (prefix: lib.hasPrefix prefix path) keepPrefixes) ./.;
  targets = [ "Main" ];
  srcdir = "./src";
}
