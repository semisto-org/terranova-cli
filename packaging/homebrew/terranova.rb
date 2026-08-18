# Formule Homebrew de terranova (ISC-426). Vit dans le tap
# semisto-org/homebrew-tap (Formula/terranova.rb).
# Installation : brew install semisto-org/tap/terranova
class Terranova < Formula
  desc "Le compagnon CLI de Terranova (app.semisto.org) — projets, compta, botanique, pépinière"
  homepage "https://github.com/semisto-org/terranova-cli"
  version "0.5.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.5.0/terranova_darwin_arm64"
      sha256 "af3c50c9b1dec768968c6d254b56529c0987eb6707bf84b242a0939bec6ec978"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.5.0/terranova_darwin_amd64"
      sha256 "2441da0637d4be7b87489593948b885cf2ed96634810d6aa178c10986846df09"
    end
  end
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.5.0/terranova_linux_arm64"
      sha256 "4101d44ab873f6716def42aa003f0dc6fd72365543a3afe8fe302e20284e56ed"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.5.0/terranova_linux_amd64"
      sha256 "7a15135c762afd909c03ca27c7e2ed2c68e5d5c48e0ceedea3691937afd93d41"
    end
  end

  def install
    binary = Dir["terranova_*"].first
    bin.install binary => "terranova"
  end

  test do
    assert_match "0.5.0", shell_output("#{bin}/terranova version --quiet")
  end
end
