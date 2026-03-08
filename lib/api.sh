_catalog_emoji=(
    "🇻🇳"  # vietsub
    "🎤"   # voice-over / thuyết minh
    "🔊"   # dubbed / lồng tiếng
)

_fetch_search_results() {
    local encoded_query="$1"
    local i catalog_id catalog_label emoji url

    for i in "${!catalog_ids[@]}"; do
        catalog_id="${catalog_ids[$i]}"
        catalog_label="${catalog_labels[$i]}"
        emoji="${_catalog_emoji[$i]:-🎬}"
        url="$BASE_URL/catalog/movie/$catalog_id/search=$encoded_query.json"

        curl -fsSL "$url" 2>/dev/null \
            | jq -r --arg label "$catalog_label" --arg emoji "$emoji" \
                '.metas[]? | [.id, .name, ((.year // .releaseInfo // "") | tostring), $label, $emoji] | @tsv' 2>/dev/null
    done
}

_search_cache_path() {
    local query="$1"
    printf '/tmp/vnstream_search_%s_%s' "$(printf '%s' "$query" | md5sum | cut -d' ' -f1)" "$$"
}

_format_movie_lines() {
    awk -F'\t' '!seen[$1]++ {
        n=NR; name=$2; year=$3
        if (year != "")
            printf "%d\t🎬  %s  (%s)\n", n, name, year
        else
            printf "%d\t🎬  %s\n", n, name
    }'
}

# Pick phim: nếu chưa có cache → stream fetch vào fzf + cache;
#            nếu đã có cache → load từ cache vào fzf
pick_movie() {
    local query="$1"
    local cache_file encoded_query
    local raw fzf_exit=0
    local -a out=()
    local key typed_query selected picked_key

    cache_file="$(_search_cache_path "$query")"
    encoded_query="$(printf '%s' "$query" | jq -sRr @uri)"

    if [[ -s "$cache_file" ]]; then
        # Có cache → load từ file, không fetch
        raw="$( {
            _format_movie_lines < "$cache_file"
            printf 'b\t← Quay lại tìm kiếm\n'
        } | choose_with_fzf_event '🎬 ' \
            "Kết quả: \"$query\"  ·  Enter chọn  ·  Alt-Enter tìm mới  ·  Esc quay lại")" || return 1
    else
        # Chưa có cache → stream fetch vào fzf + lưu cache
        fzf_exit=0
        raw="$( {
            _fetch_search_results "$encoded_query" \
                | tee "$cache_file" \
                | _format_movie_lines
            printf 'b\t← Quay lại tìm kiếm\n'
        } | fzf \
            "${_FZF_BASE_FLAGS[@]}" \
            --prompt '🎬 ' \
            --header "⏳  Đang tìm \"$query\"..." \
            --bind "load:change-header:Kết quả: \"$query\"  ·  Enter chọn  ·  Alt-Enter tìm mới  ·  Esc quay lại" \
            --print-query \
            --expect=enter,alt-enter
        )" || fzf_exit=$?

        if (( fzf_exit != 0 )); then
            [[ -s "$cache_file" ]] || rm -f "$cache_file"
            if (( fzf_exit == 130 || fzf_exit == 2 )); then
                return 1
            fi
        fi
    fi

    mapfile -t out <<< "$raw"
    typed_query="${out[0]:-}"
    key="${out[1]:-}"
    selected="${out[2]:-}"

    if [[ -z "$key" && -z "$selected" && -n "$typed_query" ]]; then
        key="enter"
    fi

    if [[ "$key" == "alt-enter" || ( "$key" == "enter" && -z "$selected" ) ]]; then
        typed_query="$(sanitize_field "$typed_query")"
        [[ -n "$typed_query" ]] || return 1
        printf 'SEARCH:%s\n' "$typed_query"
        return 0
    fi

    [[ -n "$selected" ]] || return 1

    picked_key="${selected%%$'\t'*}"
    [[ "$picked_key" == "b" ]] && return 1

    sed -n "${picked_key}p" "$cache_file"
}

fetch_streams() {
    local movie_id="$1"

    curl -fsSL "$BASE_URL/stream/movie/$movie_id.json" 2>/dev/null \
        | jq -r '
            .streams[]?
            | [
                (.url // (if .infoHash then ("magnet:?xt=urn:btih:" + .infoHash) else "" end)),
                (.name // ""),
                (.description // "")
              ]
            | @tsv
        ' 2>/dev/null
}

resolve_movie_id_by_name() {
    local movie_name="$1"
    local encoded_query line id name
    local fallback_id=""
    local wanted_name lower_name

    wanted_name="$(sanitize_field "$movie_name")"
    lower_name="${wanted_name,,}"
    encoded_query="$(printf '%s' "$wanted_name" | jq -sRr @uri)"

    while IFS= read -r line; do
        IFS=$'\t' read -r id name _ _ _ <<< "$line"
        [[ -z "$id" ]] && continue

        [[ -z "$fallback_id" ]] && fallback_id="$id"
        if [[ "${name,,}" == "$lower_name" ]]; then
            printf '%s\n' "$id"
            return 0
        fi
    done < <(_fetch_search_results "$encoded_query")

    [[ -n "$fallback_id" ]] || return 1
    printf '%s\n' "$fallback_id"
}
