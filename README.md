# آی‌روم (IRoom)

**پلتفرم کلاس آنلاین متن‌باز برای کاربران فارسی‌زبان**

[![Go](https://img.shields.io/badge/Go-1.24-blue)](https://go.dev)
[![SvelteKit](https://img.shields.io/badge/SvelteKit-5-orange)](https://svelte.dev)
[![TailwindCSS](https://img.shields.io/badge/TailwindCSS-4-cyan)](https://tailwindcss.com)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

---

## معرفی

آی‌روم یک پلتفرم کلاس آنلاین متن‌باز است که مشابه اسکای‌روم و بلوجینز طراحی شده و کاملاً فارسی با پشتیبانی RTL است. این پلتفرم برای اجرا روی سرورهای کم‌منبع (۱ هسته / ۱ گیگابایت رم) بهینه‌سازی شده و تا ۱۰۰ کاربر همزمان را پشتیبانی می‌کند.

## ویژگی‌ها

- **ویدیو و صدا**: اتصال WebRTC از طریق LiveKit SFU
- **اشتراک‌گذاری صفحه**: اشتراک‌گذاری صفحه نمایش
- **تخته‌سفید**: وایت‌بورد تعاملی با Fabric.js و همگام‌سازی آنی
- **گفتگوی متنی**: چت آنلاین با WebSocket
- **ضبط جلسه**: ضبط سمت مرورگر با MediaRecorder API
- **مدیریت کلاس**: ایجاد کلاس، ثبت‌نام دانش‌آموز، زمان‌بندی جلسات
- **پنل مدیریت**: مدیریت کاربران، کلاس‌ها، جلسات و تنظیمات
- **API خارجی**: اتصال به سیستم‌های LMS/CMS از طریق REST API
- **تقویم جلالی**: پشتیبانی از تقویم فارسی

## فناوری‌ها

| لایه | فناوری |
|------|--------|
| بک‌اند | Go + Echo + SQLite WAL |
| فرانت‌اند | SvelteKit + TailwindCSS RTL |
| ویدیو | LiveKit SFU (sidecar) |
| وایت‌بورد | Fabric.js |
| ریورس پراکسی | Cady |
| استقرار | Docker Compose |

## نصب سریع

### با Docker (توصیه شده)

```bash
git clone https://github.com/iroom/iroom.git
cd iroom
cp config.yaml config.yaml.bak
# ویرایش config.yaml با تنظیمات خود
docker-compose up -d
```

### نصب دستی

**پیش‌نیازها:**
- Go 1.24+
- Node.js 20+
- LiveKit Server (اختیاری برای ویدیو)

```bash
# بک‌اند
go build -o server ./cmd/server
./server

# فرانت‌اند
cd web
npm install
npm run dev
```

## ساختار پروژه

```
iroom/
├── cmd/server/main.go          # نقطه ورود
├── internal/
│   ├── config/                 # تنظیمات YAML
│   ├── database/               # SQLite + مایگریشن‌ها
│   ├── handlers/               # API handlers
│   ├── middleware/              # JWT, CORS, نقش‌ها
│   ├── models/                 # مدل‌های داده
│   ├── repository/             # queries پایگاه داده
│   ├── services/               # LiveKit token generation
│   └── pkg/                    # ابزارهای مشترک
├── web/src/
│   ├── lib/                    # API client, stores, components
│   └── routes/                 # صفحات SvelteKit
├── docker-compose.yml
├── Dockerfile
├── Caddyfile
└── config.yaml
```

## تنظیمات

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "iroom.db"

jwt:
  secret: "your-secret-key"
  access_expiry: 15
  refresh_expiry: 10080

livekit:
  api_key: "devkey"
  api_secret: "secret"
  url: "ws://localhost:7880"

external:
  api_key: "your-external-api-key"
```

## API

### احراز هویت
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/auth/me
```

### کلاس‌ها
```
GET|POST    /api/v1/classes
PUT|DELETE  /api/v1/classes/:id
POST        /api/v1/classes/:id/enroll
GET         /api/v1/classes/:id/students
```

### جلسات
```
GET|POST    /api/v1/sessions
GET         /api/v1/sessions/:id
POST        /api/v1/sessions/:id/start|end
POST        /api/v1/sessions/:id/livekit-token
POST        /api/v1/sessions/:id/recordings
```

### API خارجی (با کلید API)
```
POST /api/v1/external/users
POST /api/v1/external/classes
POST /api/v1/external/sessions
GET  /api/v1/external/status
GET  /api/v1/external/stats
```

## نقش‌ها

| نقش | دسترسی |
|-----|--------|
| **مدیر** | مدیریت کامل سیستم |
| **مدرس** | ایجاد کلاس/جلسه، شروع/پایان جلسه |
| **دانش‌آموز** | مشاهده کلاس‌ها، شرکت در جلسات |

## پیش‌نیاز سخت‌افزاری

- **حداقل**: ۱ هسته CPU / ۱ گیگابایت RAM
- **توصیه شده**: ۲ هسته CPU / ۲ گیگابایت RAM
- **حافظه مورد نیاز**:
  - Go Backend: ~۳۰-۵۰ MB
  - LiveKit SFU: ~۱۰۰-۱۲۰ MB
  - SQLite: ~۵-۱۰ MB
  - Caddy: ~۱۰-۱۵ MB

## مجوز

MIT License
