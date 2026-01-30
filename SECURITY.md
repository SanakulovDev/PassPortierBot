# Security Policy

Biz **PassPortierBot** xavfsizligiga jiddiy e'tibor qaratamiz. Ushbu hujjat sizga xavfsizlik siyosatimiz va zaifliklarni qanday xabar qilish haqida ma'lumot beradi.

## 📦 Supported Versions

Faqat oxirgi barqaror versiya xavfsizlik yangilanishlarini oladi.

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## 🛡 Technical Security Details

PassPortierBot foydalanuvchi ma'lumotlarini himoya qilish uchun quyidagi texnologiyalardan foydalanadi:

### 1. Zero-Knowledge Architecture

- **Master Key** (Asosiy Kalit) **hech qachon** doimiy xotirada (HDD/SSD) yoki bazada saqlanmaydi.
- Kalit faqat sessiya davomida **RAM (Tezkor Xotira)** da saqlanadi.
- Server o'chirilganda yoki qayta ishga tushirilganda, barcha kalitlar yo'qoladi.

### 2. Encryption (Shifrlash)

- Barcha parollar **AES-256-GCM** algoritmi yordamida shifrlanadi.
- Har bir yozuv uchun unikal **Nonce** ishlatiladi.

### 3. Auto-Delete (Avto-O'chirish)

- Foydalanuvchi yuborgan maxfiy xabarlar (parollar) bot tomonidan qabul qilingandan so'ng **darhol o'chiriladi**.
- Bu Telegram chat tarixida maxfiy ma'lumotlar qolmasligini ta'minlaydi.

## 🐛 Reporting a Vulnerability

Agar siz loyihada xavfsizlik zaifligini topsangiz, iltimos, omma oldida oshkor qilmang. Bizga quyidagi manzil orqali xabar bering:

- **Email**: sanakulov.dev@gmail.com
- **Telegram**: [@Sanakulov_dev](https://t.me/Sanakulov_dev)

Biz zaiflikni tekshirib, tez orada tuzatishga harakat qilamiz.
