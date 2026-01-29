# 🔐 PassPortierBot

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?style=flat&logo=redis)](https://redis.io/)
[![Gemini AI](https://img.shields.io/badge/AI-Gemini_2.0-8E75B2?style=flat&logo=google)](https://deepmind.google/technologies/gemini/)

**PassPortierBot** — bu Telegram orqali ishlaydigan **Zero-Knowledge** parollar menejeri. U sizning login va parollaringizni **Gemini AI** yordamida matn, rasm yoki ovozli xabarlardan ajratib oladi va eng yuqori xavfsizlik standartlari asosida shifrlab saqlaydi.

---

## 🚀 Asosiy Xususiyatlar

- **🔒 Zero-Knowledge Arxitekturasi**: Sizning **Master Key**ingiz (Asosiy Kalit) hech qachon bazaga yozilmaydi. U faqat vaqtinchalik RAM (tezkor xotira)da saqlanadi.
- **🧠 Multimodal AI Tahlil**:
  - 📝 **Matn**: Login/parol yuboring, bot tushunadi.
  - 📸 **Rasm (OCR)**: Ekranni rasmga olib yuboring (skrinshot).
  - 🎙 **Ovoz (Speech-to-Text)**: "Loginim admin, parolim 12345" deb ayting.
- **🔐 Harbiy Darajadagi Shifrlash**: Barcha ma'lumotlar **AES-256-GCM** algoritmi bilan shifrlanadi.
- **⏳ Xavfsiz Sessiya**: `/unlock` qilingandan so'ng, kalit RAMda **30 daqiqa** turadi va keyin avtomatik o'chib ketadi.

---

## 🛠 Texnologik Stack

- **Til**: Go (Golang) 1.24
- **AI**: Google Gemini 2.0 Flash (`google.golang.org/genai`)
- **Bazalar**:
  - PostgreSQL (GORM bilan) — Shifrlangan ma'lumotlar uchun.
  - Redis — Kelajakdagi sessiya/kesh boshqaruvi uchun.
- **Containerization**: Docker & Docker Compose.

---

## ⚙️ O'rnatish va Ishga Tushirish

Loyihani klonlang va kerakli sozlamalarni kiriting.

### 1. Talablar

- Docker & Docker Compose
- Google Gemini API Key
- Telegram Bot Token

### 2. O'rnatish

```bash
# Loyihani yuklab oling
git clone https://github.com/SanakulovDev/PassPortierBot.git
cd PassPortierBot

# .env faylni yarating
cp .env.example .env
```

`.env` faylini ochib o'z ma'lumotlaringizni kiriting:

```ini
BOT_TOKEN=sizning_bot_tokeningiz
GEMINI_API_KEY=sizning_gemini_kalitingiz
DB_USER=admin
DB_PASSWORD=secret
DB_NAME=passportier
```

### 3. Ishga tushirish (Docker)

```bash
# Barcha servislarni ko'tarish
make pro
# Yoki
docker-compose up -d --build
```

---

## 📖 Foydalanish Qo'llanmasi

### 1️⃣ Sessiyani Ochish (`/unlock`)

Bot ishlashi uchun avval unga Master Kalitni (32 ta belgi) berish kerak. Bu kalit faqat RAMda saqlanadi.

```
/unlock bu_juda_uzun_va_maxfiy_kalit_32x
```

✅ _Javob: "🔓 Sessiya ochildi!..."_

### 2️⃣ Ma'lumot Saqlash (AI)

Sessiya ochilgandan so'ng, xohlagan formatda ma'lumot yuboring.

- **Matn**: "Facebook loginim: ali, parolim: 123456"
- **Rasm**: Login/parol yozilgan qog'oz yoki ekran rasmi.
- **Ovoz**: Ovozli xabarda login va parolni ayting.

AI buni tahlil qilib, **JSON** formatga o'tkazadi va shifrlab bazaga saqlaydi.

### 3️⃣ Parolni Olish (`/get`)

Kerakli xizmat nomini yozing:

```
/get facebook
```

✅ _Bot shifrlangan parolni ochib beradi._

---

## 🛡 Xavfsizlik Tafsilotlari

1. **Master Key**: Hech qayerda (diskda, loglarda, bazada) saqlanmaydi. Agar server o'chsa, kalit ham yo'qoladi.
2. **AES-256-GCM**: Har bir yozuv alohida shifrlanadi.
3. **GenAI Privacy**: Ma'lumotlar tahlil uchun Google serveriga yuboriladi, lekin saqlanmaydi (Google AI siyosatiga qarang).

---

## 👨‍💻 Muallif

Loyihani yaratuvchi: **[SanakulovDev](https://github.com/SanakulovDev)**
Litsenziya: MIT
