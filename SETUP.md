# راهنمای نصب و اجرای آی‌روم

## پیش‌نیازها

### 1. Go (نسخه ۱.۲۴ یا بالاتر)
```bash
# دانلود و نصب Go
cd /tmp
wget https://mirrors.aliyun.com/golang/go1.24.6.linux-amd64.tar.gz
mkdir -p ~/go-sdk
tar -C ~/go-sdk -xzf go.tar.gz
echo 'export PATH=$PATH:$HOME/go-sdk/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 2. Node.js (نسخه ۲۰ یا بالاتر)
```bash
# اگر nvm دارید:
nvm install 20
nvm use 20
node --version
```

### 3. LiveKit Server (برای ویدیو/صدا)
```bash
# دانلود LiveKit
curl -sSL https://get.livekit.io | bash

# یا نصب دستی
wget https://github.com/livekit/livekit/releases/latest/download/livekit_linux_amd64.tar.gz
tar xzf livekit_linux_amd64.tar.gz
sudo mv livekit-server /usr/local/bin/

# ساخت فایل تنظیمات
cat > livekit.yaml << 'EOF'
port: 7880
rtc:
  tcp_port: 7881
  port_range_start: 50000
  port_range_end: 60000
  use_external_ip: false
keys:
  devkey: secret
logging:
  level: info
EOF
```

---

## اجرای پروژه

### مرحله ۱: راه‌اندازی بک‌اند (Go)
```bash
cd ~/StudioProjects/iroom

# ساخت باینری
export PATH=$PATH:$HOME/go-sdk/go/bin
GOTOOLCHAIN=local go build -o server ./cmd/server

# اجرای سرور
./server
# → سرور روی پورت ۸۰۸۰ اجرا می‌شود
```

### مرحله ۲: راه‌اندازی فرانت‌اند (SvelteKit)
```bash
cd ~/StudioProjects/iroom/web

# نصب وابستگی‌ها
npm install

# اجرای سرور توسعه
npm run dev
# → فرانت‌اند روی پورت ۵۱۷۳ اجرا می‌شود
```

### مرحله ۳: راه‌اندازی LiveKit (برای ویدیو/صدا)
```bash
# در ترمینال جدید
cd ~/StudioProjects/iroom
livekit-server --config livekit.yaml
# → LiveKit روی پورت ۷۸۸۰ اجرا می‌شود
```

---

## ورود به سیستم

1. مرورگر را باز کنید: `http://localhost:5173`
2. ایمیل: `admin@iroom.local`
3. رمز عبور: `admin123`

---

## Docker (جایگزین نصب دستی)

```bash
cd ~/StudioProjects/iroom

# ساخت فایل .env
cp .env.example .env

# اجرای همه سرویس‌ها
docker-compose up -d

# مشاهده لاگ‌ها
docker-compose logs -f
```

---

## عیب‌یابی

| مشکل | راه حل |
|-------|--------|
| `خطا در اتصال به سرور` | مطمئن شوید سرور Go روی پورت ۸۰۸۰ اجراست |
| `could not establish signal connection` | LiveKit اجرا نیست. آن را اجرا کنید |
| `UI بدون استایل` | فایل `app.css` در `+layout.svelte` ایمپورت شده؟ |
| `کلاس‌ها نمایش داده نمی‌شوند` | سرور را ری‌استارت کنید و دوباره لاگین کنید |
| `پورت ۵۱۷۳ اشغال است` | `pkill -f vite` و دوباره `npm run dev` |
