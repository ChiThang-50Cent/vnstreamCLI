pick_search_keyword() {
    local -a options=()
    local line hidx query selected selected_key
    local -a valid_keywords=() out=()

    ensure_storage_files

    hidx=0
    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        valid_keywords+=("$line")
        options+=("h:${hidx}${TAB}🕐  $(sanitize_field "$line")")
        (( hidx++ )) || true  # (( 0 )) = falsy, guard against set -e
    done < <(head -n 20 "$SEARCH_HISTORY_FILE")
    options+=("b${TAB}← Quay lại")

    local raw fzf_exit=0
    raw="$(printf '%s\n' "${options[@]}" | fzf \
        "${_FZF_BASE_FLAGS[@]}" \
        --print-query \
        --prompt '🔍 Tìm kiếm: ' \
        --header 'Gõ từ khóa rồi Enter  ·  hoặc chọn lịch sử  ·  Esc quay lại')" || fzf_exit=$?

    # 130 = Esc/abort → quay lại, 2 = error
    if (( fzf_exit == 130 || fzf_exit == 2 )); then
        return 1
    fi
    # fzf_exit 0 = chọn item, fzf_exit 1 = no match (nhưng query vẫn output)

    mapfile -t out <<< "$raw"
    query="${out[0]:-}"     # --print-query: luôn là line đầu
    selected="${out[1]:-}"  # item được chọn (rỗng nếu không chọn gì)

    if [[ -n "$selected" ]]; then
        selected_key="${selected%%$'\t'*}"
        case "$selected_key" in
            b) return 1 ;;
            h:*)
                hidx="${selected_key#h:}"
                printf '%s\n' "${valid_keywords[$hidx]}"
                return 0
                ;;
        esac
    fi

    # Không chọn item → dùng typed query
    query="$(sanitize_field "$query")"
    [[ -n "$query" ]] || return 1
    printf '%s\n' "$query"
}

show_all_links() {
    local movie_name="$1"
    shift
    local -a all_streams=("$@")
    local -a options=()
    local idx final_link sname sdesc

    for idx in "${!all_streams[@]}"; do
        IFS=$'\t' read -r final_link sname sdesc <<< "${all_streams[$idx]}"
        [[ -z "$final_link" ]] && continue
        if [[ -n "$sdesc" ]]; then
            options+=("x${TAB}📋  $((idx + 1)))  $sname  ·  $sdesc")
        else
            options+=("x${TAB}📋  $((idx + 1)))  $sname")
        fi
    done
    options+=("b${TAB}← Quay lại")

    printf '%s\n' "${options[@]}" | choose_with_fzf '📋 ' \
        "Tất cả link: $movie_name" --no-sort >/dev/null || true
}

select_stream_loop() {
    local movie_line="$1"
    local movie_id movie_name
    local -a all_streams=() stream_menu=() out=()
    local input idx final_link sname sdesc stream_name line
    local raw key typed_query selected selected_key
    local fzf_exit=0
    local tmpfile

    IFS=$'\t' read -r movie_id movie_name _year _label <<< "$movie_line"
    tmpfile="$(mktemp /tmp/vnstream_streams.XXXXXX)"

    # Lần đầu: stream trực tiếp vào fzf (không flash terminal)
    raw="$( {
        fetch_streams "$movie_id" \
            | tee "$tmpfile" \
            | awk -F'\t' '{
                n=NR; sname=$2; sdesc=$3
                if (sdesc != "")
                    printf "%d\t▶  %s  ·  %s\n", n, sname, sdesc
                else
                    printf "%d\t▶  %s\n", n, sname
            }'
        printf 'all\t📋  Xem tất cả link\n'
        printf 'b\t← Quay lại kết quả\n'
    } | fzf \
        "${_FZF_BASE_FLAGS[@]}" \
        --prompt '▶  ' \
        --header "⏳  Đang tải stream: $movie_name..." \
        --bind "load:change-header:Stream: $movie_name  ·  Enter chọn  ·  Alt-Enter tìm mới  ·  Esc quay lại" \
        --print-query \
        --expect=enter,alt-enter
    )" || fzf_exit=$?

    # Load data từ tmpfile vào array để loop tiếp
    mapfile -t all_streams < "$tmpfile"
    rm -f "$tmpfile"

    if (( fzf_exit == 130 || fzf_exit == 2 )); then
        return 1
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

    input="${selected%%$'\t'*}"

    if (( ${#all_streams[@]} == 0 )); then
        show_notice "Không có stream nào cho: $movie_name"
        return 1
    fi

    # Xử lý lần chọn đầu tiên
    while true; do
        if [[ "$input" == "b" ]]; then
            return 1
        fi

        if [[ "$input" == "all" ]]; then
            show_all_links "$movie_name" "${all_streams[@]}"
        elif [[ "$input" =~ ^[0-9]+$ ]] && (( input >= 1 && input <= ${#all_streams[@]} )); then
            IFS=$'\t' read -r final_link stream_name _ <<< "${all_streams[$((input - 1))]}"
            if [[ -z "$final_link" ]]; then
                show_notice "Stream này không có link phát."
            else
                save_watched "$movie_name" "$stream_name" "$final_link" "$movie_id"
                if ! play_in_vlc "$final_link" "$movie_name" "$stream_name"; then
                    show_notice "Không mở được VLC hoặc phát thất bại."
                fi
            fi
        fi

        # Loop: hiện lại menu stream từ array đã load
        stream_menu=()
        for idx in "${!all_streams[@]}"; do
            IFS=$'\t' read -r _link sname sdesc <<< "${all_streams[$idx]}"
            if [[ -n "$sdesc" ]]; then
                stream_menu+=("$((idx + 1))${TAB}▶  $sname  ·  $sdesc")
            else
                stream_menu+=("$((idx + 1))${TAB}▶  $sname")
            fi
        done
        stream_menu+=("all${TAB}📋  Xem tất cả link")
        stream_menu+=("b${TAB}← Quay lại kết quả")

        if ! raw="$(printf '%s\n' "${stream_menu[@]}" | choose_with_fzf_event '▶  ' \
            "Stream: $movie_name  ·  ${#all_streams[@]} nguồn  ·  Enter chọn  ·  Alt-Enter tìm mới  ·  Esc quay lại")"; then
            return 1
        fi

        mapfile -t out <<< "$raw"
        typed_query="${out[0]:-}"
        key="${out[1]:-}"
        selected="${out[2]:-}"

        if [[ "$key" == "alt-enter" || ( "$key" == "enter" && -z "$selected" ) ]]; then
            typed_query="$(sanitize_field "$typed_query")"
            [[ -n "$typed_query" ]] || continue
            printf 'SEARCH:%s\n' "$typed_query"
            return 0
        fi

        selected_key="${selected%%$'\t'*}"
        input="$selected_key"
    done
}

search_flow() {
    local query="${1:-}"
    local movie_line stream_result

    [[ -n "$query" ]] || return 1

    while true; do
        save_history "$query"

        # pick_movie tự handle cache: lần đầu stream+cache, lần sau load từ cache
        while true; do
            if ! movie_line="$(pick_movie "$query")"; then
                return 1  # Back từ movie list → quay về Home
            fi

            if [[ "$movie_line" == SEARCH:* ]]; then
                query="${movie_line#SEARCH:}"
                break
            fi

            if stream_result="$(select_stream_loop "$movie_line")"; then
                if [[ "$stream_result" == SEARCH:* ]]; then
                    query="${stream_result#SEARCH:}"
                    break
                fi
            fi

            # Back từ stream → quay lại movie picker (từ cache)
        done
    done
}

clear_search_history() {
    : > "$SEARCH_HISTORY_FILE"
}

clear_watched_history() {
    : > "$WATCH_HISTORY_FILE"
}

home_menu_pick() {
    local -a watched=() history_keywords=() valid_history=() options=()
    local idx count ts movie_name stream_name short_ts
    local line hidx
    local raw key typed_query selected selected_key
    local -a out=()

    ensure_storage_files
    mapfile -t history_keywords < <(head -n 20 "$SEARCH_HISTORY_FILE")
    mapfile -t watched < <(tail -n 30 "$WATCH_HISTORY_FILE" | tac)
    count="${#watched[@]}"

    if (( ${#history_keywords[@]} > 0 )); then
        options+=("sep_h${TAB}─── Tìm kiếm gần đây (${#history_keywords[@]}) ───────────────────")
        hidx=0
        for line in "${history_keywords[@]}"; do
            [[ -z "$line" ]] && continue
            valid_history+=("$line")
            options+=("h:${hidx}${TAB}🕐  $(sanitize_field "$line")")
            (( hidx++ )) || true
        done
    else
        options+=("sep_h${TAB}─── Tìm kiếm gần đây ───────────────────────────")
        options+=("noop_h${TAB}   Chưa có từ khóa tìm kiếm")
    fi

    if (( count > 0 )); then
        options+=("sep_w${TAB}─── Đã xem ($count) ──────────────────────────")
        for idx in "${!watched[@]}"; do
            line="${watched[$idx]}"
            IFS=$'\t' read -r ts movie_name stream_name _ <<< "$line"
            [[ -z "$movie_name" ]] && continue
            # Shorten: "2026-03-03 18:29:04" -> "03/03 18:29"
            short_ts="${ts:8:2}/${ts:5:2} ${ts:11:5}"
            options+=("watch:${idx}${TAB}🎬  [$short_ts]  $movie_name  ·  $stream_name")
        done
    else
        options+=("sep_w${TAB}─── Đã xem ──────────────────────────────────")
        options+=("noop_w${TAB}   Chưa có video nào được xem")
    fi

    options+=("sep_a${TAB}─── Hành động ─────────────────────────────────")
    options+=("act_clear_watched${TAB}🗑  Xóa lịch sử xem")
    options+=("act_clear_search${TAB}🗑  Xóa lịch sử tìm kiếm")
    options+=("act_exit${TAB}🚪  Thoát")

    raw="$(printf '%s\n' "${options[@]}" | choose_with_fzf_event '❯ ' \
        'VNStream  ·  ↑↓ điều hướng  ·  / lọc  ·  Esc thoát' \
        --no-sort)" || return 1

    mapfile -t out <<< "$raw"
    typed_query="${out[0]:-}"
    key="${out[1]:-}"
    selected="${out[2]:-}"

    if [[ "$key" == "alt-enter" || ( "$key" == "enter" && -z "$selected" ) ]]; then
        typed_query="$(sanitize_field "$typed_query")"
        [[ -n "$typed_query" ]] || return 1
        printf 'SEARCH:%s\n' "$typed_query"
        return 0
    fi

    selected_key="${selected%%$'\t'*}"
    if [[ "$selected_key" == h:* ]]; then
        printf 'SEARCH:%s\n' "${valid_history[${selected_key#h:}]}"
        return 0
    fi

    printf '%s\n' "$selected_key"
}

replay_watched_by_index() {
    local picked_idx="$1"
    local -a watched=()
    local line ts movie_name stream_name link movie_id
    local movie_line stream_result
    local resolve_tmp resolve_pid resolve_loading_exit=0
    local -a spinner=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")
    local spin_idx=0

    mapfile -t watched < <(tail -n 30 "$WATCH_HISTORY_FILE" | tac)
    line="${watched[$picked_idx]:-}"
    [[ -n "$line" ]] || return 1

    IFS=$'\t' read -r ts movie_name stream_name link movie_id <<< "$line"
    if [[ -z "$link" ]]; then
        show_notice "Không có link phát cho mục này."
        return 1
    fi

    if [[ -z "$movie_id" && "${VNSTREAM_REPLAY_RESOLVE_MISSING_ID:-1}" == "1" ]]; then
        resolve_tmp="$(mktemp /tmp/vnstream_resolve_id.XXXXXX)"
        (resolve_movie_id_by_name "$movie_name" 2>/dev/null || true) > "$resolve_tmp" &
        resolve_pid=$!

        {
            while kill -0 "$resolve_pid" 2>/dev/null; do
                printf 'loading\t%s  Dang tim phim...\n' "${spinner[$spin_idx]}"
                spin_idx=$(( (spin_idx + 1) % ${#spinner[@]} ))
                sleep 0.1
            done
            wait "$resolve_pid" >/dev/null 2>&1 || true
            printf 'done\t✓  Hoan tat\n'
        } | fzf \
            "${_FZF_BASE_FLAGS[@]}" \
            --prompt '  ' \
            --header "⏳  Dang tim phim: $movie_name" \
            --no-sort \
            --disabled >/dev/null 2>&1 || resolve_loading_exit=$?

        movie_id="$(<"$resolve_tmp")"
        rm -f "$resolve_tmp"

        if (( resolve_loading_exit == 130 || resolve_loading_exit == 2 )); then
            kill "$resolve_pid" >/dev/null 2>&1 || true
            return 1
        fi
    fi

    if [[ -n "$movie_id" ]]; then
        movie_line="$movie_id"$'\t'"$movie_name"$'\t\t'
        if stream_result="$(select_stream_loop "$movie_line")"; then
            if [[ "$stream_result" == SEARCH:* ]]; then
                printf '%s\n' "$stream_result"
            fi
        fi
        return 0
    fi

    show_notice "Không lấy được danh sách source, phát lại link cũ."
    if ! play_in_vlc "$link" "$movie_name" "$stream_name"; then
        show_notice "Không mở được VLC hoặc phát thất bại."
        return 1
    fi
    return 0
}

main_loop() {
    local choice idx query

    while true; do
        if ! choice="$(home_menu_pick)"; then
            exit 0
        fi

        case "$choice" in
            act_clear_watched)
                if confirm_action "Xóa toàn bộ lịch sử xem?"; then
                    clear_watched_history
                fi
                ;;
            act_clear_search)
                if confirm_action "Xóa toàn bộ lịch sử tìm kiếm?"; then
                    clear_search_history
                fi
                ;;
            act_exit)
                exit 0
                ;;
            watch:*)
                idx="${choice#watch:}"
                if query="$(replay_watched_by_index "$idx")"; then
                    if [[ "$query" == SEARCH:* ]]; then
                        query="${query#SEARCH:}"
                        [[ -n "$query" ]] && search_flow "$query" || true
                    fi
                fi
                ;;
            SEARCH:*)
                query="${choice#SEARCH:}"
                [[ -n "$query" ]] && search_flow "$query" || true
                ;;
            *)
                ;;
        esac
    done
}
