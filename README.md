# 🔐 PassPortierBot

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Security](https://img.shields.io/badge/Encryption-AES--256--GCM-green?style=flat&logo=lock)](https://en.wikipedia.org/wiki/AES-GCM)

**PassPortierBot** — Telegram orqali ishlaydigan **Zero-Knowledge** parollar menejeri.

---

## 🔒 Security Architecture

| Feature | Implementation |
|---------|---------------|
| **Encryption** | AES-256-GCM (Authenticated) |
| **Key Derivation** | Argon2id (64MB, 4 threads) |
| **Salt Strategy** | Unique 16-byte salt per encryption |
| **Session TTL** | 30 minutes (RAM only) |
| **Password Storage** | ❌ NEVER stored |

### Zero-Knowledge Design
```
/unlock password → Store passphrase in RAM (30 min TTL)
#save data       → Generate unique Salt → DeriveKey → Encrypt → Store
#get data        → DeriveKey with stored Salt → Decrypt → Verify password!
```

---

## 🚀 Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message |
| `/unlock [password]` | Open session (30 min) |
| `/list` | Show ALL saved secrets |
| `/get [service]` | Get single secret |
| `#service data` | Save/Update secret |
| `#service` | Retrieve secret |

---

## 📁 Project Structure

```
internal/
├── bot/           # Bot initialization
├── handlers/      # Command handlers
│   ├── start.go   # /start
│   ├── unlock.go  # /unlock
│   ├── get.go     # /get
│   ├── list.go    # /list
│   └── text.go    # #hash parser
├── services/      # Business logic
│   ├── auth.go    # Session management
│   ├── password.go# Save/Retrieve
│   └── secret.go  # SecretService
├── repository/    # Data access layer
│   └── secret.go  # SecretRepository
├── vault/         # Session storage (RAM)
├── crypto/        # Encryption
│   ├── manager.go # CryptoManager (Encrypt/Decrypt)
│   ├── aes.go     # Low-level AES
│   └── kdf.go     # Argon2id KDF
├── models/        # Database models
└── storage/       # DB initialization
```

---

## ⚙️ Quick Start

```bash
git clone https://github.com/SanakulovDev/PassPortierBot.git
cd PassPortierBot
cp .env.example .env
# Edit .env with your BOT_TOKEN
make pro
```

### Environment Variables

```ini
BOT_TOKEN=your_telegram_bot_token
DB_HOST=db
DB_USER=admin
DB_PASSWORD=secret
DB_NAME=passportier
DB_PORT=5432
```

---

## 📖 Usage Examples

### 1️⃣ Open Session
```
/unlock mySecretPassword
```
✅ _Session opened for 30 minutes_

### 2️⃣ Save Secret
```
#instagram mypassword123
```
🛡 _Message auto-deleted for security_

### 3️⃣ Get Secret
```
#instagram
```
⏰ _Response auto-hides in 10 seconds_

### 4️⃣ List All Secrets
```
/list
```
📋 _Shows all decrypted secrets (10s auto-hide)_

---

## 🛠 Development

```bash
make setup    # Build containers
make restart  # Restart bot
make logs     # View logs
make stop     # Stop containers
```

---

## 👨‍💻 Author

**[SanakulovDev](https://github.com/SanakulovDev)** | Built with 42.uz System Design principles

