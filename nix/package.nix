# statecast バイナリの Nix パッケージ定義（go-overlay buildGoApplication）
{ pkgs, go, src, lib, version ? "0.1.0" }:

let
  # src と modules で同じフィルタ済みソースを参照する
  source = lib.cleanSource src;
in
pkgs.buildGoApplication {
  inherit go version;

  pname = "statecast";
  src = source;
  subPackages = [ "cmd/statecast" ];
  modules = source + "/govendor.toml";

  # サンドボックス内でも統合テストは通る（Unix ソケットは /tmp 配下、daemon テストは事前ビルド済みバイナリを使用）
  doCheck = true;
}
