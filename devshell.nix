{ pkgs }:
pkgs.mkShell {
  # Add build dependencies
  packages = [ pkgs.go ];

  # Add environment variables
  env = { };

  # Load custom bash code
  shellHook = "\n";
}
