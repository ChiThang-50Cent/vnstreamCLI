# VNStream UI — Kế hoạch cải thiện TUI liên tục

## Vấn đề

Có những thời điểm **terminal thô lộ ra** thay vì TUI fzf xuyên suốt. Sau khi phân tích code, có 4 điểm gây lỗi:

| # | Điểm | File | Mô tả | Mức độ |
|---|------|------|-------|--------|
| 1 | `require_cmd` | `config.sh:27` | `echo >&2` + `exit 1` thẳng ra terminal nếu thiếu `curl`/`jq`/`fzf` | 🟡 Trung bình |
| 2 | VLC stderr leak | `playback.sh:94` | VLC đôi khi in ra stderr trước khi redirect xong | 🟢 Nhẹ |
| 3 | `resolve_movie_id_by_name` | `flow.sh:322-323` | Network call **blocking ngoài fzf** → terminal trắng vài giây | 🔴 Nghiêm trọng |
| 4 | `fetch_streams` blocking | `flow.sh:327` | Network call **blocking ngoài fzf** khi replay lịch sử → terminal trắng | 🔴 Nghiêm trọng |

---

## Chi tiết từng điểm & hướng fix

### Điểm 3 & 4 — `replay_watched_by_index` (nghiêm trọng nhất)

**Nguyên nhân gốc rễ**: Luồng hiện tại khi user chọn phim từ lịch sử "Đã xem":
```
Đóng fzf  →  resolve_movie_id (curl)  →  fetch_streams (curl)  →  Mở fzf mới
                   ↑ terminal trắng ở đây ↑
```

**Hướng fix**: Stream `fetch_streams` trực tiếp vào `select_stream_loop` (y hệt pattern đang dùng ở lần tìm kiếm mới), bỏ bước pre-fetch blocking. Hàm `resolve_movie_id_by_name` nếu cần vẫn gọi nhưng phải bọc bên trong fzf loading screen.

**Phương án cụ thể**:
- Bỏ `mapfile -t preview_streams < <(fetch_streams ...)` — đây là root cause
- Gọi thẳng `select_stream_loop "$movie_line"` nếu đã có `movie_id`
- Nếu thiếu `movie_id`: Wrap `resolve_movie_id_by_name` vào fzf loading (dùng `--header '⏳ Đang tìm phim...'` với một fzf spinner giả hoặc process substitution)

---

### Điểm 1 — `require_cmd` error screen

**Hướng fix**: Thay `echo >&2` bằng một fzf "error screen" (hoặc đơn giản hơn: dùng `dialog`/`whiptail` nếu có, fallback về styled `echo` với box border bằng ANSI).

**Phương án cụ thể**:
- Viết hàm `fatal_error <message>` trong `ui.sh`
- Dùng ANSI box đơn giản hiển thị lỗi rõ ràng, pause chờ Enter trước khi exit

---

### Điểm 2 — VLC stderr leak

**Hướng fix**: Đảm bảo redirect stderr đủ sớm trong `_launch_vlc_detached`.

**Phương án cụ thể**:
- Thêm `2>/dev/null` ngay khi exec VLC thay vì append vào log file (hoặc giữ log nhưng redirect trước khi bất kỳ output nào ra stdout/stderr của terminal hiện tại)

---

## Checklist

### Planning
- [x] Phân tích toàn bộ codebase, xác định 4 điểm terminal lộ ra ngoài TUI
- [x] Viết kế hoạch chi tiết

### Execution
- [ ] **[Prio 1]** Fix `replay_watched_by_index` (`flow.sh`): bỏ blocking `fetch_streams` pre-fetch, gọi thẳng `select_stream_loop`
- [ ] **[Prio 1]** Fix `replay_watched_by_index` (`flow.sh`): wrap `resolve_movie_id_by_name` vào fzf loading nếu thiếu `movie_id`
- [ ] **[Prio 2]** Viết hàm `fatal_error` trong `ui.sh`, cập nhật `require_cmd` trong `config.sh` dùng hàm đó
- [ ] **[Prio 3]** Fix VLC stderr leak trong `playback.sh`

### Verification
- [ ] Chọn phim từ lịch sử "Đã xem" → không thấy terminal trắng khi đang resolve/fetch
- [ ] Xóa phim khỏi lịch sử rồi replay → flow resolve `movie_id` vẫn có loading screen
- [ ] Chạy script thiếu `curl` → thấy error screen đẹp, không phải dòng chữ thô
- [ ] Phát VLC → không có text rác từ VLC in ra terminal
