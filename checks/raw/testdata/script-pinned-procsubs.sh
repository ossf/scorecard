#!/bin/bash

# both of these fetch the same immutable URL, one through a pipe and one
# through a process substitution
curl -sSL https://raw.githubusercontent.com/dotnet/install-scripts/5b142a1e445a6f060d6430b661408989e9580b85/src/dotnet-install.sh | bash /dev/stdin
bash <(curl --silent --show-error "https://raw.githubusercontent.com/rhysd/actionlint/7b75d16d41920ec126e6f3269db0c6f3ab613c38/scripts/download-actionlint.bash")
