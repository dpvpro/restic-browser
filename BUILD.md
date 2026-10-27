# Сборка и запуск

## Предварительные требования

- **Wails3 CLI** v3.x — проверка: `wails3 version`
- **Go** 1.26+ — проверка: `go version`
- **Node.js + npm** (используется Vite 7)
- **Linux:** GTK4 + WebKitGTK 6.0
  - Debian/Ubuntu: `sudo apt install libgtk-4-dev libwebkitgtk-6.0-dev`
  - Arch: `sudo pacman -S gtk4 webkitgtk-6.0`
- **restic** в `$PATH` (или путь передаётся флагом `--restic`, см. ниже)

## Запуск в режиме разработки

```bash
cd restic-browser
wails3 dev
```

Что происходит:

- собирается dev-бинарник;
- поднимается Vite dev-сервер на `http://127.0.0.1:9245`;
- открывается окно приложения;
- изменения в Go-файлах вызывают пересборку, изменения во фронтенде применяются через HMR.

Выход — `Ctrl+C`.

Эквивалент через Taskfile: `wails3 task dev`.

## Сборка

```bash
wails3 build
```

Результат — `bin/Restic-Browser`. Production-ассеты фронтенда встраиваются в бинарник.

Перед сборкой задача автоматически:

- ставит npm-зависимости;
- генерирует биндинги Go → TypeScript (`frontend/bindings`);
- собирает фронтенд (`frontend/dist`);
- генерирует иконки и `.desktop`-файл.

Запуск собранного бинарника:

```bash
./bin/Restic-Browser
./bin/Restic-Browser --restic=/usr/bin/restic
```

## Пакеты для распространения

```bash
wails3 package
```

Создаёт платформенные пакеты на основе `build/config.yml`:

- **Linux** — AppImage, deb, rpm (и файлы для AUR);
- **Windows** — NSIS/MSIX;
- **macOS** — `.app`, DMG.

Артефакты складываются в `bin/` и `build/`.

## Полезные команды

| Задача | Команда |
|---|---|
| Запустить уже собранный бинарник | `wails3 task run` |
| Только пересобрать фронтенд | `cd frontend && npm run build` |
| Проверка типов TypeScript | `cd frontend && npx tsc` |
| Перегенерировать биндинги (после изменения экспортируемых методов Go) | `wails3 generate bindings -ts -clean=true` |
| Тесты бэкенда | `go test ./backend/...` |
| Полная очистка | `rm -rf bin frontend/dist frontend/bindings` |

## Переменные окружения restic

Приложение подхватывает стандартные переменные restic, если репозиторий не выбран вручную:

```bash
RESTIC_REPOSITORY=/path/to/repo \
RESTIC_PASSWORD=secret \
wails3 dev
```

Также поддерживаются:

- `RESTIC_REPOSITORY_FILE`, `RESTIC_PASSWORD_FILE`;
- креды S3 (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`), Backblaze B2 (`B2_ACCOUNT_ID`, `B2_ACCOUNT_KEY`), Azure (`AZURE_ACCOUNT_NAME`, `AZURE_ACCOUNT_KEY`);
- префиксы репозиториев: `sftp:`, `rest:`, `rclone:`, `s3:`, `b2:`, `azure:`.

## Отладка Go-кода

В `.vscode/launch.json` есть конфигурация `Debug Restic-Browser (GO)`. По `F5` выполняется задача `wails3 build DEV=true`, затем запускается `bin/Restic-Browser` с переменными для тестового репозитория (`backend/restic/test-repo`).

## Примечания

- `wails3 dev` не пробрасывает свои аргументы приложению. Если нужен нестандартный `--restic`, используйте собранный `./bin/Restic-Browser --restic=...`.
- Биндинги в `frontend/bindings` генерируются автоматически; при изменении сигнатур экспортируемых методов Go их нужно перегенерировать (команда выше).
- В dev-режиме после полной перезагрузки страницы приложение может показать «No repository selected» — репозиторий откроется кнопкой **Open Repository** или после перезапуска `wails3 dev`.
