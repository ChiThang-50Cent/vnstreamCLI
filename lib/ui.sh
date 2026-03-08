# Catppuccin Mocha color scheme
_FZF_COLORS='fg:#cdd6f4,bg:#1e1e2e,bg+:#313244,hl:#89b4fa,fg+:#cdd6f4,hl+:#89dceb,info:#cba6f7,prompt:#89b4fa,pointer:#f38ba8,marker:#a6e3a1,spinner:#f5c2e7,header:#74c7ec,border:#45475a'
_FZF_BASE_FLAGS=(
    --layout=reverse
    --border=rounded
    --cycle
    --no-multi
    --color="$_FZF_COLORS"
    --pointer='❯'
    --delimiter=$'\t'
    --with-nth=2..
    --bind='esc:abort'
)

# choose_with_fzf <prompt> <header> [extra fzf flags...]
choose_with_fzf() {
    local prompt="$1"
    local header="$2"
    shift 2
    local selected

    selected="$({ cat; } | fzf \
        "${_FZF_BASE_FLAGS[@]}" \
        --prompt "$prompt" \
        --header "$header" \
        "$@")" || return 1

    printf '%s\n' "${selected%%$'\t'*}"
}

# choose_with_fzf_event <prompt> <header> [extra fzf flags...]
# Output (3 lines):
#   1) query    (typed text)
#   2) key      (enter|alt-enter|"")
#   3) selected (full selected row, may be empty)
choose_with_fzf_event() {
    local prompt="$1"
    local header="$2"
    shift 2

    local raw fzf_exit=0
    local -a out=()
    local key query selected

    raw="$({ cat; } | fzf \
        "${_FZF_BASE_FLAGS[@]}" \
        --prompt "$prompt" \
        --header "$header" \
        --print-query \
        --expect=enter,alt-enter \
        "$@")" || fzf_exit=$?

    if (( fzf_exit == 130 || fzf_exit == 2 )); then
        return 1
    fi

    mapfile -t out <<< "$raw"

    # fzf with --print-query --expect outputs:
    #   line1=query, line2=key (empty on default Enter), line3=selected
    query="${out[0]:-}"
    key="${out[1]:-}"
    selected="${out[2]:-}"

    if [[ -z "$key" && -z "$selected" && -n "$query" ]]; then
        key="enter"
    fi

    printf '%s\n%s\n%s\n' "$query" "$key" "$selected"
}

show_notice() {
    local message="$1"
    printf 'ok\t  ✓  OK\n' | choose_with_fzf '  ● ' "$message" >/dev/null || true
}

confirm_action() {
    local message="$1"
    local choice
    if ! choice="$(printf 'no\t  ✗  Không\nyes\t  ✓  Có\n' | choose_with_fzf '  ● ' "$message")"; then
        return 1
    fi
    [[ "$choice" == "yes" ]]
}

fatal_error() {
    local message="$1"
    local line="Missing required command"

    if [[ -n "$message" ]]; then
        line="$message"
    fi

    printf '\n\033[1;31m+--------------------------------------------------+\033[0m\n'
    printf '\033[1;31m|                    VNStream Error                |\033[0m\n'
    printf '\033[1;31m+--------------------------------------------------+\033[0m\n'
    printf '\033[1;31m| %s\033[0m\n' "$line"
    printf '\033[1;31m+--------------------------------------------------+\033[0m\n\n'

    if [[ -t 1 ]]; then
        printf 'Nhan Enter de thoat... '
        IFS= read -r _ </dev/tty || true
    fi

    exit 1
}
