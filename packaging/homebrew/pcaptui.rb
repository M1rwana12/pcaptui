class Pcaptui < Formula
  desc "Terminal interface to tshark: packet list, protocol tree and hex view, over SSH"
  homepage "https://github.com/m1rwana12/pcaptui"
  version "1.2.0"
  license "MIT"

  # The release binaries rather than a source build: they are what CI produced
  # and what the checksums in the release cover, so what Homebrew installs is
  # the same artefact that was tested and published.
  on_macos do
    on_intel do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.2.0/pcaptui_1.2.0_darwin_amd64.tar.gz"
      sha256 "3fc6094bea93d5bab712fc690a623b5cb03c9561775147226f8f5287ed96794f"
    end
    on_arm do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.2.0/pcaptui_1.2.0_darwin_arm64.tar.gz"
      sha256 "317ac3d0c8c66ad572ec96879f92c4b395a581170de7830a9112d5e783a858e7"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.2.0/pcaptui_1.2.0_linux_amd64.tar.gz"
      sha256 "2f69b54e12808c1ee7bc37581517f30fdeda70310dad8b15662be17c4cf77efa"
    end
    on_arm do
      url "https://github.com/m1rwana12/pcaptui/releases/download/v1.2.0/pcaptui_1.2.0_linux_arm64.tar.gz"
      sha256 "893c376ebc3c39913e3af1373327f3625e240cbc99ace265df43c659dde672fe"
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
