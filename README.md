# Описание
В кластере 3 контейнера: Master: Основной сервер PostgreSQL, Slave: Резервный сервер, который синхронизируется с Master, Arbiter: Сервер, используемый для определения работоспособности Master и Slave
Раз в 5 секунд Master проверяет связи до слейва и до арбитра, если отсутствуют обе связи, то происходит блокировка всех входящих подключений на Master через iptables, путём изменения политики по умолчанию на DROP.
Раз в секунду Slave проверяет связи до мастера и от арбитра к мастеру, если отсутствуют обе связи, то происходит промоут до Master, путем создания триггер файла.
Кластер инициализируется с помощью bash-скриптов, а также присваивание кластерного айпи-адреса.
# Цель
Проверить количество потерянных записей при различных параметрах синхронизации транзакций в случае изменения мастера. 
# Запуск
Настроить config и запустить создание контейнеров:
```bash
docker-compose up -d
```
# Тестирование
1. Ошибка мастера
На мастере происходит какая-то ошибка и он становится недоступен, на slave происходит promote до master и данные начинают поступать напрямую в slave. Это можно воспроизвести с помощью блокирования ip адреса мастера с помощью iptables)

При `synchronous_commit = off` произошла потеря 47 записи из 1 миллиона.

При `synchronous_commit = remote_apply` потери данных не обнаружено.

2. Ошибка slave

При `synchronous_commit = off` потери данных не обнаружено.

При `synchronous_commit = remote_apply` потери данных не обнаружено.
