{ lib, yarn, stdenv, mkYarnPackage, libsass, python, nodejs, fetchurl }:
with builtins;

let
  inherit (lib) hasPrefix splitString concatMapStrings;

  filterSourcePrefixes = root: prefixes:
    let keepPrefixes = (map (pa: toString pa) prefixes);
    in filterSource
    (path: type: (any (prefix: hasPrefix prefix path) keepPrefixes)) root;

  nodeHeaders = fetchurl {
    url =
      "https://nodejs.org/download/release/v${nodejs.version}/node-v${nodejs.version}-headers.tar.gz";
    sha256 = "0a8frg7848lw4wjl8l0c4c125wb4m982jwpfs4crbh5qzjmlr8bn";
  };

  deps = mkYarnPackage {
    name = "scylla-frontend-dependencies";
    src = filterSourcePrefixes ./. [ ./package.json ./yarn.lock ./default.nix ];
    packageJson = ./package.json;
    yarnLock = ./yarn.lock;
    # yarnNix = ./yarn.nix;
    publishBinsFor = [ "cross-env" "webpack" ];
    pkgConfig = {
      node-sass = {
        buildInputs = [ libsass python ];
        postInstall = ''
          node scripts/build.js --tarball=${nodeHeaders}
        '';
      };
    };
  };
in stdenv.mkDerivation {
  name = "scylla-frontend";
  phases = [ "buildPhase" ];
  nativeBuildInputs = [ deps yarn ];
  src = filterSourcePrefixes ./. [
    ./build
    ./config
    ./index.html
    ./package.json
    ./src
    ./static
    ./test
    ./webpack.config.js
    ./yarn.lock
    ./.eslintrc.js
    ./.eslintignore
    ./.babelrc
    ./.prettierrc
  ];
  buildPhase = ''
    mkdir -p $out
    cp -r $src $out/tmp
    chmod -R 0777 $out/tmp
    cd $out/tmp
    ln -sf ${deps}/node_modules/libexec/scylla/node_modules node_modules
    yarn run build
    cp -r dist/* $out
    rm -rf $out/tmp
  '';
}
