sources = ["env:///bin/packages", "https://github.com/cashapp/hermit-packages.git", "https://github.com/foxygoat/hermit-packages.git"]
env = {
  GOBIN: "${HERMIT_ENV}/out/bin",
  PATH: "${GOBIN}:${PATH}",
}
manage-git = false
