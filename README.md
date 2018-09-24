# Nix Flavoured Continuous Integration

Scylla is a simple CI server that solves one thing:
Evaluate [Nix](https://nixos.org/nix/) derivations and inform you and GitHub
about the results.

Scylla is implemented in Go, and needs nothing but Nix for building, logging,
and caching.

I try to keep the moving parts as reliable as possible, since at the end of the
day, all we care about is that it works.

## What Scylla Can Do For You

* Create binaries
* Run tests
* Populate your Nix cache
* Update your GitHub PR status
* Serve logs of your project builds

## Getting Started

1. Get an OAuth token

   Navigate to [https://github.com/settings/tokens](https://github.com/settings/tokens)
   and generate a new OAuth token. It only needs the <code>repo:status</code> permission.

2. Add the webhook

   Go to `https://github.com/$owner/$repo/settings/hooks` (substitute your `owner`/`repo` in the URL).

   Add a webhook that points to your server, like `https://$host/github-webhook` (substitute `host` here to the location of your server, you can also use something like [ngrok](http://ngrok.com/) for trying it out).  
   These settings are required:
   * Content type: `application/json`
   * Secret: (anything you want)
   * Enable at least the `pull request` event. The rest will at the momemt be simply ignored.

3. Configure the server

   Configuration is done via Environment variables (although flags also work).
   You need to set the following:
   
    * `GITHUB_TOKEN`: The token you created in the first step
    * `GITHUB_USER`: The token you created in the first step

4. Building the server

       nix build -f ci.nix scylla

5. Running the server

       ./result/bin/scylla

6. Add a `ci.nix` file to the project you want to use it with.

   The following is just an example of the structure. It's strongly recommended
   that you use a pinned version of `nixpkgs` so both Scylla and you are
   actually building identical things.

       { nixpkgs ? import (fetchGit {
         url = https://github.com/NixOS/nixpkgs;
         ref = "24429d66a3fa40ca98b50cad0c9153e80f56c4a2";
       }) {} }: {
         app-binary = callPackage ./. {};
         app-tests = recurseIntoAttrs {
           callPackage ./. { slowTests = true; };
           callPackage ./. { fastTests = true; };
         }
       }

   All atributes in the returned attrset will be evaluated, in this case
   `app-binary` and `app-tests`.
   What Scylla actually does is to simply call `nix-build` on your `ci.nix` and
   store the results.
   
   So you can make sure the `ci.nix` is working by doing that yourself first locally.
   
   The `recurseIntoAttrs` function can be used to also build nested attrsets.
   Otherwise only functions in the top-level will be built.



## TODO

### Must have

- [ ] Resume aborted builds when server restarts
- [ ] Handle build timeouts better
- [ ] Remove old builds automatically
- [ ] Better scheduling, right now it's limited by number of Cores
- [ ] Option to restart builds easily
- [ ] Cancel still running builds for PRs that are updated

### For 1.0
- [ ] Safely execute actions depending on build result (probably a webhook of sorts?)
- [ ] When docker containers are built, push them to a registry

### Nice to have
- [ ] Some better support for indicating tests
- [ ] Build everything into a profile, so comparing generations would be possible
