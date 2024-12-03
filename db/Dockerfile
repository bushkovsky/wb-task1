# Используем официальный образ PostgreSQL
FROM postgres:15

# Устанавливаем переменные окружения для автоматической настройки
ENV POSTGRES_USER=ibushkovskij
ENV POSTGRES_PASSWORD=psg
ENV POSTGRES_DB=orders

# Копируем файл init.sql в директорию, из которой PostgreSQL будет его исполнять при запуске
COPY init.sql /docker-entrypoint-initdb.d/

# Указываем порт для PostgreSQL
#docker exec -it ordersWb /bin/bash
EXPOSE 5432
#docker run --name ordersWb -d -p 5432:5432 orderswb
#docker exec -it ordersWb netstat -an | grep 5432
#docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' ordersWb