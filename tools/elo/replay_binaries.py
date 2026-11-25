"""Verified Go 1.26.2 darwin/arm64 rebuilds after the Git history rewrite."""

# Archived match SHA-256 -> clean-history rebuild of the same engine configuration.
REBUILT_SHA256 = {
    # Base1850, 91aeb4e
    "1df5969e78f2ee42a7f980fb4b5043426dd928f527847e7af9cbe763bb8364ab":
        "e665be19443f65362511486995955658f06cd9bca1b60dbadcefb81b02037fec",
    # Base2200, a43a8a3
    "bae8740966bc6cc3b4bcf3b708a3f1f06b63a1a48667274b43d58807e84f1a95":
        "ff4749fe461ab76b7ff383a6180e34d48bb2e15b00c24673b923ef876446648c",
    # RankingOnly, 04efa19 with ranking-only.patch and the quietverify build tag
    "1323dd9cc0c1c8eafa0b311f53e3b69314acd91004db3648a8c5f0a7e6ff93e7":
        "f2a2806e4901c28513755ade3580624ac00c3f89f0c2cd6d7bd008141adf58b3",
}


def matches_recorded_binary(actual, archived):
    return actual == archived or actual == REBUILT_SHA256.get(archived)
