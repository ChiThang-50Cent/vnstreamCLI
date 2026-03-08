BASE_URL="https://addon.vnstream.io.vn/guest"
DATA_DIR="${HOME}/.vnstream"
SEARCH_HISTORY_FILE="${DATA_DIR}/search_history"
WATCH_HISTORY_FILE="${DATA_DIR}/watched_history"
LEGACY_HISTORY_FILE="${HOME}/.cache/vnstream_history"
VLC_XDG_CONFIG_HOME="${DATA_DIR}/vlc_config"
VLC_XDG_CACHE_HOME="${DATA_DIR}/vlc_cache"

VLC_WIDTH="400"
VLC_HEIGHT="300"
TAB=$'\t'

catalog_ids=(
    "vnstream-vietsub"
    "vnstream-voice-over"
    "vnstream-dubbed"
)

catalog_labels=(
    "Vietsub"
    "Thuyet minh"
    "Long tieng"
)

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        if declare -F fatal_error >/dev/null 2>&1; then
            fatal_error "Missing required command: $1"
        fi
        printf 'Missing required command: %s\n' "$1" >&2
        exit 1
    fi
}

sanitize_field() {
    local value="$1"
    value="${value//$'\t'/ }"
    value="${value//$'\n'/ }"
    printf '%s\n' "$value"
}
