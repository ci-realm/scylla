with import ./nixpkgs.nix;

with builtins;
with lib;

let
  inherit (lib) hasPrefix splitString concatMapStrings;

  goDeps = stdenv.mkDerivation {
    name = "goDeps";
    src = ../Gopkg.lock;
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

in stdenv.mkDerivation {
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
}
