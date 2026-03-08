# VNStream CLI (Go)

- [English](#english)
- [Tiếng Việt](#tiếng-việt)

Terminal UI app to search VNStream movie catalogs and play selected streams in VLC.

## English

### What this app does

VNStream is a Go-based terminal app for:

- Searching VNStream movie catalogs (Subtitled / Voice-over / Dubbed)
- Browsing stream options for a selected movie
- Launching playback in VLC (`vlc` or `qvlc`)

### Requirements

- Go 1.22+ (only needed when building from source)
- VLC installed for playback (`vlc` or `qvlc` in `PATH`)
- Internet access to VNStream addon API

### Install (prebuilt release)

Current prebuilt support is **linux_amd64 only**.

```bash
curl -fsSL https://raw.githubusercontent.com/ChiThang-50Cent/vnstreamCLI/main/install.sh | bash
```

Optional installer vars:

- `VERSION` (default: `latest`)
- `INSTALL_DIR` (default: `~/.local/bin`)

### Run

```bash
vnstream
```

### Quick search

```bash
vnstream "movie name"
```

### Build from source (unsupported OS/arch)

If your platform is not `linux/amd64`, build locally:

```bash
git clone https://github.com/ChiThang-50Cent/vnstreamCLI.git
cd vnstreamCLI
go build -o vnstream .
./vnstream
```

### Notes / limitations

- Installer currently allows prebuilt install only on `linux/amd64`.
- App stores local data in `~/.vnstream` (search history, watched history, VLC config/cache).
- Without VLC, app can run but cannot play streams.

---

## Tiếng Việt

### Ứng dụng này làm gì

VNStream là ứng dụng terminal viết bằng Go, dùng để:

- Tìm phim từ catalog VNStream (Vietsub / Thuyết minh / Lồng tiếng)
- Duyệt danh sách stream của phim đã chọn
- Mở phát bằng VLC (`vlc` hoặc `qvlc`)

### Yêu cầu

- Go 1.22+ (chỉ cần khi build từ source)
- Cài VLC để phát (`vlc` hoặc `qvlc` có trong `PATH`)
- Có kết nối mạng tới VNStream addon API

### Cài đặt (bản phát hành prebuilt)

Hiện tại prebuilt chỉ hỗ trợ **linux_amd64**.

```bash
curl -fsSL https://raw.githubusercontent.com/ChiThang-50Cent/vnstreamCLI/main/install.sh | bash
```

Biến tùy chọn khi cài:

- `VERSION` (mặc định: `latest`)
- `INSTALL_DIR` (mặc định: `~/.local/bin`)

### Chạy ứng dụng

```bash
vnstream
```

### Tìm nhanh

```bash
vnstream "tên phim"
```

### Build từ mã nguồn (OS/arch chưa hỗ trợ prebuilt)

Nếu máy bạn không phải `linux/amd64`, hãy build trực tiếp:

```bash
git clone https://github.com/ChiThang-50Cent/vnstreamCLI.git
cd vnstreamCLI
go build -o vnstream .
./vnstream
```

### Ghi chú / giới hạn hiện tại

- Script cài đặt hiện chỉ cho cài prebuilt trên `linux/amd64`.
- Dữ liệu cục bộ nằm ở `~/.vnstream` (lịch sử tìm kiếm, lịch sử xem, cấu hình/cache VLC).
- Không có VLC thì app vẫn chạy nhưng không phát được stream.
