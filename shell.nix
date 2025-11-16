let
  home = builtins.getEnv "HOME";

  pkgs = (import "${home}/nix" { }).pkgs;
in
pkgs.mkShell {
  packages = with pkgs; [
    onnxruntime.dev # 包含头文件
    go
    libopus
    pkg-config
  ];
  # https://github.com/NixOS/nixpkgs/pull/347526/files
  env = {
    LD_LIBRARY_PATH = "${pkgs.stdenv.cc.cc.lib}/lib";
    CGO_ENABLED = 1;
    SILERO_MODEL_PATH = "./models/silero_vad.onnx";
    CGO_CFLAGS = "-O2 -Wno-cpp";
    CONFIG_PATH = ".config.yaml";
  };
}
