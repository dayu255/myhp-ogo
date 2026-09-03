{
  description = "Development environment template";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor =
        system:
        import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
    in
    {
      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            # Dev-Tools
            packages = with pkgs; [
              go_1_27
            ];

            # Build-Tools and Runtime
            nativeBuildInputs = with pkgs; [
              # gcc
            ];

            # Libraries needed at build
            buildInputs = with pkgs; [
              # zlib
            ];

            # Dynamic-Link Libraries needed at run
            # LD_LIBRARY_PATH = pkgs.lib.makeLibraryPath (with pkgs; [
            #   stdenv.cc.cc.lib
            # ]);

            # Environment Variables
            # MY_CAT = "cute";

            shellHook = ''
              # Shell Command when load this flake

            '';
          };
        }
      );
    };
}
