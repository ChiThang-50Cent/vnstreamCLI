ensure_storage_files() {
    mkdir -p "$DATA_DIR" "$VLC_XDG_CONFIG_HOME" "$VLC_XDG_CACHE_HOME"

    if [[ ! -f "$SEARCH_HISTORY_FILE" && -f "$LEGACY_HISTORY_FILE" ]]; then
        cp "$LEGACY_HISTORY_FILE" "$SEARCH_HISTORY_FILE"
    fi

    touch "$SEARCH_HISTORY_FILE" "$WATCH_HISTORY_FILE"
}

save_history() {
    local query="$1"
    ensure_storage_files
    awk -v q="$query" 'BEGIN { print q } $0 != q' "$SEARCH_HISTORY_FILE" | head -n 20 > "${SEARCH_HISTORY_FILE}.tmp"
    mv "${SEARCH_HISTORY_FILE}.tmp" "$SEARCH_HISTORY_FILE"
}

save_watched() {
    local movie_name stream_name link movie_id timestamp
    movie_name="$(sanitize_field "$1")"
    stream_name="$(sanitize_field "$2")"
    link="$(sanitize_field "$3")"
    movie_id="$(sanitize_field "${4:-}")"

    ensure_storage_files

    if [[ -n "$link" ]]; then
        awk -F'\t' -v target_link="$link" '$4 != target_link' "$WATCH_HISTORY_FILE" > "${WATCH_HISTORY_FILE}.tmp"
        mv "${WATCH_HISTORY_FILE}.tmp" "$WATCH_HISTORY_FILE"
    fi

    timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    printf '%s\t%s\t%s\t%s\t%s\n' "$timestamp" "$movie_name" "$stream_name" "$link" "$movie_id" >> "$WATCH_HISTORY_FILE"
    tail -n 200 "$WATCH_HISTORY_FILE" > "${WATCH_HISTORY_FILE}.tmp"
    mv "${WATCH_HISTORY_FILE}.tmp" "$WATCH_HISTORY_FILE"
}

clear_search_history() {
    : > "$SEARCH_HISTORY_FILE"
}

clear_watched_history() {
    : > "$WATCH_HISTORY_FILE"
}
