{ stdenv, lib, buildGoPackage, fetchFromGitHub, makeWrapper, runCommand, remarshal, scylla-frontend }:

with builtins;
with lib;

rec {
  runDir = runCommand "scylla-dir" {} ''
    mkdir -p $out
    ln -s ${scylla-frontend} $out/public
  '';

  scylla-bin = buildGoPackage rec {
    name = "scylla-unstable-${version}";
    version = "2018-07-23";

    goPackagePath = "github.com/manveru/scylla";

    keepPrefixes = (map (pa: toString pa) [ ./Makefile ./queue ./server ]);
    src = filterSource (path: type:
      (hasSuffix ".go" path) ||
      (any (prefix: lib.hasPrefix prefix path) keepPrefixes)) ./.;

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
