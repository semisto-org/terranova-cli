# Formule Homebrew de terranova (ISC-426). Vit dans le tap
# semisto-org/homebrew-tap (Formula/terranova.rb).
# Installation : brew install semisto-org/tap/terranova
class Terranova < Formula
  desc "Le compagnon CLI de Terranova (app.semisto.org) — projets, compta, botanique, pépinière"
  homepage "https://github.com/semisto-org/terranova-cli"
  version "0.4.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.4.0/terranova_darwin_arm64"
      sha256 "d27934cffa49cfed4767e3b121a1751626955855ce636e681dc65ef40738f6e5"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.4.0/terranova_darwin_amd64"
      sha256 "287fdb443b33232b4fe6d0525c0966abb92d20979b9c1be21212896dc47721f3"
    end
  end
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.4.0/terranova_linux_arm64"
      sha256 "df178e4f82558a58f892fd4b0a366e1d8184db27b59554485feef265235ddda8"
    else
      url "https://github.com/semisto-org/terranova-cli/releases/download/v0.4.0/terranova_linux_amd64"
      sha256 "3552585680316c9c16afb7131b41ccfaa8c8f62a25a2b085ffe9da67bd29b604"
    end
  end

  def install
    binary = Dir["terranova_*"].first
    bin.install binary => "terranova"
  end

  test do
    assert_match "0.4.0", shell_output("#{bin}/terranova version --quiet")
  end
end
