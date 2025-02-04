class Kutelog < Formula
  desc "A tool for logging with WebSocket support"
  homepage "https://github.com/{{ .Owner }}/{{ .Name }}"
  version "{{ .Version }}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/{{ .Owner }}/{{ .Name }}/releases/download/{{ .Tag }}/kutelog-darwin-arm64"
      sha256 "{{ .DarwinArm64SHA }}"
    else
      url "https://github.com/{{ .Owner }}/{{ .Name }}/releases/download/{{ .Tag }}/kutelog-darwin-amd64"
      sha256 "{{ .DarwinAmd64SHA }}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/{{ .Owner }}/{{ .Name }}/releases/download/{{ .Tag }}/kutelog-linux-arm64"
      sha256 "{{ .LinuxArm64SHA }}"
    else
      url "https://github.com/{{ .Owner }}/{{ .Name }}/releases/download/{{ .Tag }}/kutelog-linux-amd64"
      sha256 "{{ .LinuxAmd64SHA }}"
    end
  end

  def install
    bin.install Dir["kutelog-*"].first => "kutelog"
  end

  test do
    system "#{bin}/kutelog", "--version"
  end
end
