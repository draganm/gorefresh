{ pkgs }:
pkgs.mkShell {
  # Add build dependencies
  packages = [ pkgs.go_1_23 ];

  # Add environment variables
  env = { };

  # Load custom bash code
  shellHook = "\n";
}
