class Pcaptui < Formula
  desc "Terminal interface to tshark: packet list, protocol tree and hex view, over SSH"
  homepage "https://github.com/m1rwana12/pcaptui"
  version "1.0.0"
  license "MIT"

  # The release binaries rather than a source build: they are what CI produced
  # and what the checksums in the release cover, so what Homebrew installs is
  # the same artefact that was tested and published.
  on_macos do
    on_intel do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.0.0/pcaptui_1.0.0_darwin_amd64.tar.gz"
      sha256 "eab6c16486e65991a91e08823dda1bba7647fdcd4ee46168ba8b626026835158"
    end
    on_arm do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.0.0/pcaptui_1.0.0_darwin_arm64.tar.gz"
      sha256 "ac0774fd076182bdbbe715536063c4751b81e6acf5650c88de0144af6c67bc29"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.0.0/pcaptui_1.0.0_linux_amd64.tar.gz"
      sha256 "4d0a55041a019fcae683c0da389a8279d992ee610fafc93a722ceee9fa60235a"
    end
    on_arm do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.0.0/pcaptui_1.0.0_linux_arm64.tar.gz"
      sha256 "e1e75076d4af1d1d3851d3daf25775d280439ffdb1642dc73c688ceaa6a42488"
    end
  end

  # Every dissector, display filter and statistic comes from tshark; without it
  # this program has nothing to show and says so and stops.
  depends_on "wireshark"

  def install
    bin.install "pcaptui"
    doc.install "README.md", "docs/UserGuide.md", "docs/FAQ.md"
  end

  test do
    # --pass-thru=false is required, not decoration. When stdout is not a
    # terminal - which it never is under brew test - pcaptui hands its
    # arguments to tshark, so a bare --version prints tshark's version and this
    # would assert nothing about the formula. Measured, not assumed.
    assert_match "pcaptui", shell_output("#{bin}/pcaptui --pass-thru=false --version")

    # A capture it can read end to end, through the pass-thru path, which is
    # what runs when stdout is not a terminal. It exercises the tshark this
    # formula just installed rather than only checking that a binary exists.
    (testpath/"empty.pcap").write \
      "\xd4\xc3\xb2\xa1\x02\x00\x04\x00\x00\x00\x00\x00\x00\x00\x00\x00" \
      "\xff\xff\x00\x00\x01\x00\x00\x00".b
    system bin/"pcaptui", "-r", testpath/"empty.pcap", "--pass-thru=true"
  end
end
