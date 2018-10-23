{ stdenv, lib, buildGoPackage, fetchFromGitHub, makeWrapper, runCommand, remarshal }:

with builtins;
with lib;

rec {
  inherit (lib) hasPrefix splitString concatMapStrings;

  goDeps = stdenv.mkDerivation {
    name = "goDeps";
    src = ./Gopkg.lock;
    phases = "buildPhase";
    buildInputs = [ remarshal ];
    buildPhase = ''
      remarshal --indent-json -if toml -i $src -of json -o $out
    '';
  };

  fixUrl = name:
    if (hasPrefix "golang.org" name) then
      "https://go.googlesource.com/" + (elemAt (splitString "/" name) 2)
    else
      if (hasPrefix "google.golang.org" name) then
        "https://github.com/golang/" + (elemAt (splitString "/" name) 1)
      else
        "https://" + name;

  projects = (fromJSON (readFile goDeps.out)).projects;

  mkProject = project:
    stdenv.mkDerivation {
      name = replaceStrings ["/"] ["-"] project.name;

      src = fetchGit {
        url = fixUrl project.name;
        rev = project.revision;
      } // (if project?branch then { ref = project.branch; } else {});

      phases = [ "buildPhase" ];

      buildPhase = ''
        mkdir -p $out/package
        cp -r $src/* $out/package
        echo "${project.name}" > $out/name
      '';
    };

  projectSources = map mkProject projects;

  depTree = stdenv.mkDerivation {
    name = "depTree";

    src = projectSources;

    phases = [ "buildPhase" ];

    buildPhase = ''
      mkdir -p $out
      for pkg in $src; do
        echo building "$pkg"
        name="$(cat $pkg/name)"
        mkdir -p "$out/vendor/$name"
        cp -r $pkg/package/* "$out/vendor/$name"
      done
    '';
  };

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
