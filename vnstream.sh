#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
trap 'rm -f /tmp/vnstream_search_*_"$$"' EXIT

# shellcheck source=lib/config.sh
source "$SCRIPT_DIR/lib/config.sh"
# shellcheck source=lib/storage.sh
source "$SCRIPT_DIR/lib/storage.sh"
# shellcheck source=lib/ui.sh
source "$SCRIPT_DIR/lib/ui.sh"
# shellcheck source=lib/api.sh
source "$SCRIPT_DIR/lib/api.sh"
# shellcheck source=lib/playback.sh
source "$SCRIPT_DIR/lib/playback.sh"
# shellcheck source=lib/flow.sh
source "$SCRIPT_DIR/lib/flow.sh"

require_cmd curl
require_cmd jq
require_cmd fzf
ensure_storage_files

if [[ $# -gt 0 ]]; then
    search_flow "$*" || true
    exit 0
fi

main_loop
