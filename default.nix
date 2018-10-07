{ stdenv, lib, buildGoPackage, fetchFromGitHub, makeWrapper, runCommand }:

with builtins;
with lib;

rec {
  runDir = runCommand "scylla-dir" {} ''
    mkdir -p $out
    ln -s ${./public} $out/public
    ln -s ${./templates} $out/templates
  '';

  scylla-bin = buildGoPackage rec {
    name = "scylla-unstable-${version}";
    version = "2018-07-23";
    rev = "277ad49d97dd0861b889ee7a0d8922f4549affe4";

    goPackagePath = "github.com/manveru/scylla";

    keepPrefixes = (map (pa: toString pa) [ ./Makefile ./queue ]);
    src = filterSource (path: type:
      (hasSuffix ".go" path) ||
      (any (prefix: hasPrefix prefix path) keepPrefixes)) ./.;

    goDeps = ./deps.nix;

    preBuild = ''
      go generate ${goPackagePath}
      # don't run DB tests yet...
      go test ${goPackagePath}
    '';

    meta = {
      description = "A simple, easy to deploy Nix Continous Integration server";
      homepage = https://github.com/manveru/scylla;
      license = licenses.mit;
      maintainers = [ maintainers.manveru ];
      platforms = platforms.unix;
    };
  };

  scylla = runCommand "scylla-dir" { buildInputs = [ makeWrapper ]; } ''
    mkdir -p $out/bin
    cp ${scylla-bin}/bin/scylla $out/bin
    wrapProgram $out/bin/scylla --run "cd ${runDir}"
  '';
}
