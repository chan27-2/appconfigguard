class Appconfigguard < Formula
  desc "A Go-based CLI tool for safely managing Azure App Configuration"
  homepage "https://github.com/chan27-2/appconfigguard"
  url "https://github.com/chan27-2/appconfigguard/releases/download/v#{version}/appconfigguard_#{version}_Darwin_#{Hardware::CPU.arch}.tar.gz"
  sha256 ""  # Will be filled by goreleaser when publishing
  license "MIT"
  version "0.1.0"  # Update this when creating actual releases

  def install
    bin.install "appconfigguard"
  end

  test do
    system "#{bin}/appconfigguard", "--help"
  end
end
