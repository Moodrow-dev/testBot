
# testBot(название прорабатывается) - Бот для учебы

## 📝 Описание
Бот предназначен для автоматизации управления учебными чатами:
- Смена режимов недели (ЧИСЛ/ЗНАМ)
- Управление списками пользователей
- Пинги участников
- Кастомизация названий чатов

## ⚙️ Функционал

### 👨‍💻 Команды для администраторов
| Команда | Описание |
|---------|----------|
| `/init` | Инициализация бота |
| `/changeweek` | Смена недели ЧИСЛ/ЗНАМ |
| `/changetitle` | Смена названия чата |
| `/setusers` | Установка списка пользователей |
| `/setmainthread` | Назначение основного чата для уведомлений |
| `/Tolstobrow` | Включение/выключение напоминания о паре

### 👤 Команды для пользователей
| Команда | Описание |
|---------|----------|
| `/ping` | Пинг всех пользователей |

## 🛠 Технологии
- Язык: Go 1.20+
- Библиотека: [telego](https://github.com/mymmrac/telego)
- Хранение данных: SQLite

## 🚀 Установка и запуск

1. Клонировать репозиторий:
```bash
git clone https://github.com/yourusername/your-bot.git
cd your-bot
```

2. Установить зависимости:
```bash
go mod download
```

3. Настроить окружение:
```bash
cp .env.example .env
# Заполнить TELEGRAM_BOT_TOKEN в .env
```

4. Запустить бота:
```bash
go run main.go
```

## 🔧 Конфигурация

Файл `.env` должен содержать:
```ini
TELEGRAM_BOT_TOKEN=ваш_токен_бота
DB_PATH=./bot.db
```

## 📦 Развертывание

### Работа без развертывания

Telegram: @MoodrowTestBot

### Docker
```bash
docker build -t week-bot .
docker run -d --env-file .env week-bot
```

### Systemd (для сервера)
```ini
[Unit]
Description=Week Management Bot
After=network.target

[Service]
User=botuser
WorkingDirectory=/opt/week-bot
ExecStart=/opt/week-bot/bot
Restart=always
Environment=TELEGRAM_BOT_TOKEN=ваш_токен

[Install]
WantedBy=multi-user.target
```

## 📬 Контакты   
Telegram: @moodroow

### Особенности оформления:
1. **Четкое разделение** команд для админов и пользователей
2. **Подробные инструкции** по установке и настройке
3. **Несколько вариантов** развертывания (Docker, systemd)
4. **Табличное представление** команд для удобства чтения
5. **Минимальные требования** к системе
