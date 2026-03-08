clamp_vlc_window_if_oversize() {
    local vlc_pid="$1"
    local i win_id width height key val target_width target_height

    [[ "${XDG_SESSION_TYPE:-}" == "x11" ]] || return 0

    for ((i = 0; i < 40; i++)); do
        if command -v xdotool >/dev/null 2>&1; then
            win_id="$(xdotool search --onlyvisible --pid "$vlc_pid" 2>/dev/null | head -n 1)"
            if [[ -z "$win_id" ]]; then
                win_id="$(xdotool search --onlyvisible --class vlc 2>/dev/null | head -n 1)"
            fi

            if [[ -n "$win_id" ]]; then
                width=""
                height=""
                while IFS='=' read -r key val; do
                    case "$key" in
                        WIDTH) width="$val" ;;
                        HEIGHT) height="$val" ;;
                    esac
                done < <(xdotool getwindowgeometry --shell "$win_id" 2>/dev/null)

                if [[ -n "$width" && -n "$height" ]]; then
                    target_width="$width"
                    target_height="$height"
                    (( target_width > VLC_WIDTH )) && target_width="$VLC_WIDTH"
                    (( target_height > VLC_HEIGHT )) && target_height="$VLC_HEIGHT"
                    if (( target_width != width || target_height != height )); then
                        xdotool windowsize "$win_id" "$target_width" "$target_height" >/dev/null 2>&1 || true
                    fi
                    return 0
                fi
            fi
        elif command -v wmctrl >/dev/null 2>&1; then
            while IFS= read -r line; do
                set -- $line
                win_id="$1"
                if [[ "$3" == "$vlc_pid" ]]; then
                    width="$6"
                    height="$7"
                    target_width="$width"
                    target_height="$height"
                    (( target_width > VLC_WIDTH )) && target_width="$VLC_WIDTH"
                    (( target_height > VLC_HEIGHT )) && target_height="$VLC_HEIGHT"
                    if (( target_width != width || target_height != height )); then
                        wmctrl -i -r "$win_id" -e "0,-1,-1,${target_width},${target_height}" >/dev/null 2>&1 || true
                    fi
                    return 0
                fi
            done < <(wmctrl -lpG 2>/dev/null)
        fi
        sleep 0.2
    done
    return 1
}

play_in_vlc() {
    local link="$1"
    local movie_name="${2:-}"
    local stream_name="${3:-}"
    local stream_title=""

    movie_name="$(sanitize_field "$movie_name")"
    stream_name="$(sanitize_field "$stream_name")"
    stream_title="$movie_name - $stream_name"
    stream_title="${stream_title//$'\r'/ }"
    local -a vlc_flags=(
        --play-and-exit
        --no-fullscreen
        --embedded-video
        --autoscale
        --no-qt-video-autoresize
        --zoom=1
        --width="$VLC_WIDTH"
        --height="$VLC_HEIGHT"
        --avcodec-hw=none
    )

    if [[ -n "$movie_name" && -n "$stream_name" ]]; then
        vlc_flags+=("--meta-title=$stream_title")
    elif [[ -n "$stream_name" ]]; then
        vlc_flags+=("--meta-title=$stream_name")
    elif [[ -n "$movie_name" ]]; then
        vlc_flags+=("--meta-title=$movie_name")
    fi
    local vlc_pid=0

    _launch_vlc_detached() {
        local vlc_cmd="$1"

        XDG_CONFIG_HOME="$VLC_XDG_CONFIG_HOME" XDG_CACHE_HOME="$VLC_XDG_CACHE_HOME" \
            nohup "$vlc_cmd" "${vlc_flags[@]}" "$link" </dev/null >/dev/null 2>/dev/null &
        vlc_pid=$!

        clamp_vlc_window_if_oversize "$vlc_pid" >/dev/null 2>&1 || true
        return 0
    }

    if command -v qvlc >/dev/null 2>&1; then
        _launch_vlc_detached qvlc
        return 0
    fi

    if command -v vlc >/dev/null 2>&1; then
        _launch_vlc_detached vlc
        return 0
    fi

    return 1
}
