# Formule Homebrew de terranova (ISC-426). À poser dans le repo
# semisto-org/homebrew-tap (Formula/terranova.rb) — création du tap :
#   gh repo create semisto-org/homebrew-tap --public
# puis : brew install semisto-org/tap/terranova
class Terranova < Formula
  desc "Le compagnon CLI de Terranova (app.semisto.org) — projets, compta, botanique, pépinière"
  homepage "https://github.com/semisto-org/terranova-cli"
  version "0.3.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.3.0/terranova_darwin_arm64"
      sha256 "bc2466851f0a8a5b365c0e9bbe2e4334b63ea82cd14d5facf70174c720a4fb2c"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.3.0/terranova_darwin_amd64"
      sha256 "25ed26f344d96dff56009255378d56fc79fab658024b326b9bff09846acd9125"
    end
  end
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.3.0/terranova_linux_arm64"
      sha256 "6bfe63d6ce1e72b762298c8af08a46b577d411df0031e93ad09e3cb637c99620"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.3.0/terranova_linux_amd64"
      sha256 "9aaa11d3df6dde881b0770b6ad015b04d1e6b49b0c007931413e535c8970bdc7"
    end
  end

  def install
    binary = Dir["terranova_*"].first
    bin.install binary => "terranova"
  end

  test do
    assert_match "0.3.0", shell_output("#{bin}/terranova version --quiet")
  end
end
