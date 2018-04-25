### Teachbot ###
Telegram task sending bot.


#### Установка ####
В консоли на сервере
```bash
# установка базы данных
sudo apt -y install postgresql-9.5
sudo -u postgresql psql  
```

И там дальше в появившейся консоли БД
```postgresql
CREATE ROLE teachbot WITH LOGIN PASSWORD 'qwerty';
CREATE DATABASE teachbot WITH OWNER teachbot;
```
Пароль qwerty можно поменять на какой угодно,
только не забыть его потом в конфигах поменять.


Далее заливаем наш архив со всем нужным на сервер.
Из-под линукса и мака это можно сделать командой scp, 
в винде есть ВНЕЗАПНО WinSCP.


В консольке на сервере создаём директорию,
например, `~/teachbot/`,
где положим файлики бота, и распаковываем там архив,
например так:
```bash
mkdir ~/teachbot
cd ~/teachbot
tar xzf /path/to/archive.tgz 
```

Переходим в директорию бота и настраиваем там 
в файлах коннект к БД. И в файле config.toml 
настраиваем access token бота, который нам даёт телеграм.
Потому что я криворукий, БД надо прописать в двух местах:
* ~/teachbot/dist/config.toml
* ~/teachbot/dist/db/dbconf.toml

Впрочем, если пароль qwerty, то менять ничего не надо

Инициализируем БД
```bash
cd ~/teachbot/dist
./goose up
```

Далее, устанавливаем tmux
```bash
apt install tmux
```

Запускаем бот
```bash
cd ~/teachbot/dist
tmux new
./runner
```

и нажимаем Ctrl+B, D, чтобы бот крутился в фоне.