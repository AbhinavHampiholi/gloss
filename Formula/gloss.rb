class Gloss < Formula
  desc "Attach verbose, commit-pinned context to git commits via git notes"
  homepage "https://github.com/AbhinavHampiholi/gloss"
  license "MIT"

  # Stable release — requires a tagged release and its matching tarball sha256.
  # To bump: change url to the new tag, then run
  #   curl -sL <url> | shasum -a 256
  # and replace the sha256 below.
  url "https://github.com/AbhinavHampiholi/gloss/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"

  # `brew install --HEAD AbhinavHampiholi/tap/gloss` builds straight from main.
  # Useful during development before any release is cut.
  head "https://github.com/AbhinavHampiholi/gloss.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"git-gloss", ldflags: "-s -w")
  end

  test do
    # --help exits 0 and prints the banner. Good enough smoke test.
    assert_match "git-gloss", shell_output("#{bin}/git-gloss --help")
  end
end
