{ pkgs }:
pkgs.buildGoModule {
  pname = "gorefresh";
  version = "0";
  src = ./.;
  vendorHash = "sha256-nr/y7WSKeXByeml/oWlIZou4p9PoRaZbEM7JhZdJD90=";
}
