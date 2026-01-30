# 🔐 PassPortierBot

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)

**PassPortierBot** — bu Telegram orqali ishlaydigan **Zero-Knowledge** parollar menejeri. U sizning login va parollaringizni xavfsiz shifrlab saqlaydi.

---

## 🚀 Asosiy Xususiyatlar

- **🔒 Zero-Knowledge Arxitekturasi**: Sizning **Master Key**ingiz (Asosiy Kalit) hech qachon bazaga yozilmaydi. U faqat vaqtinchalik RAM (tezkor xotira)da saqlanadi.
- **🛡 Xavfsizlik va Maxfiylik**:
  - � **Avto-O'chirilish**: Siz yuborgan parollar bot tomonidan darhol o'chiriladi. Tarixda qolmaydi.
  - 🔐 **AES-256-GCM**: Barcha ma'lumotlar harbiy darajadagi shifrlash bilan himoyalanadi.
- **⚡️ Tezkor va Qulay**:
  - ✍️ **Manual Kiritish**: `#service password` yoki yangi qator bilan kiritish.
  - ⏳ **Vaqtinchalik Sessiya**: `/unlock` qilingandan so'ng, kalit RAMda **30 daqiqa** turadi.

---

## 🛠 Texnologik Stack

- **Til**: Go (Golang) 1.24
- **Bazalar**: PostgreSQL (GORM bilan)
- **Containerization**: Docker & Docker Compose

---

## ⚙️ O'rnatish va Ishga Tushirish

### 1. Talablar

- Docker & Docker Compose
- Telegram Bot Token

### 2. O'rnatish

```bash
git clone https://github.com/SanakulovDev/PassPortierBot.git
cd PassPortierBot
cp .env.example .env
```

`.env` faylini to'ldiring:

```ini
BOT_TOKEN=sizning_bot_tokeningiz
DB_USER=admin
DB_PASSWORD=secret
DB_NAME=passportier
```

### 3. Ishga tushirish (Docker)

```bash
make pro
# Yoki
docker-compose up -d --build
```

---

## 📖 Foydalanish Qo'llanmasi

### 1️⃣ Sessiyani Ochish (`/unlock`)

Bot ishlashi uchun avval unga Master Kalitni bering (RAMda saqlanadi):

```
/unlock maxfiy_so'z
```

✅ _Javob: "🔓 Sessiya ochildi!..."_

### 2️⃣ Ma'lumot Saqlash

Ma'lumotni `#` belgisi bilan boshlang. Space yoki yangi qator bilan ajratishingiz mumkin:

**Oddiy usul:**

```
#instagram parol123
```

**Ko'p qatorli usul:**

```
#instagram
bu_juda_uzun_parol_yoki_kalit
```

🛡 _Siz yuborgan xabar xavfsizlik uchun darhol o'chiriladi!_

### 3️⃣ Parolni Olish

Xizmat nomini `#` bilan yozing yoki `/get` ishlating:

```
#instagram
# yoki
/get instagram
```

✅ _Bot shifrlangan parolni ochib beradi._

---

## 👨‍💻 Muallif

**[SanakulovDev](https://github.com/SanakulovDev)**
