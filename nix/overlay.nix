let
  yarn2nixSource = fetchTarball {
    url =
      "https://github.com/moretea/yarn2nix/archive/780e33a07fd821e09ab5b05223ddb4ca15ac663f.tar.gz";
    sha256 = "1f83cr9qgk95g3571ps644rvgfzv2i4i7532q8pg405s4q5ada3h";
  };

  cachixSource = fetchTarball { url = "https://cachix.org/api/v1/install"; };

in self: super: {
  go = super.go_1_12;
  tagFromGit = self.git-info "git rev-parse --verify HEAD";
  mkShell = super.mkShell.override { stdenv = self.stdenvNoCC; };
  cachix = (import cachixSource { }).cachix;
  nodejs = super.nodejs-slim-10_x;
  yarn = super.yarn.override { nodejs = self.nodejs; };
  inherit (import yarn2nixSource { nodejs = self.nodejs; })
    yarn2nix mkYarnModules mkYarnPackage;
}
