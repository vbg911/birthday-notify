# birthday-notify

Тестовое задание для стажировки в цифровых активах "Газпром-Медиа Холдинга".

Требуется написать сервис для поздравления с днем рождения.

## Запуск

### Запуск используя docker 

1. Клонировать репозиторий:

```bash
   git clone https://github.com/vbg911/birthday-notify.git
```
2. Заполнить файл `config/config.yaml` указав свои данные smtp

3. Указать в `deployments/docker-compose.yaml` переменную ENV SMTP_PASS паролем от почты

4. Перейти в директроию deployments:

```bash
   cd deployments
```

5. Запустить приложение в docker:

```bash
   docker compose up
```

## Примеры запросов
### В postman можно посмотреть примеры запросов и их краткое описание
[<img src="https://run.pstmn.io/button.svg" alt="Run In Postman" style="width: 128px; height: 32px;">](https://god.gw.postman.com/run-collection/24125419-f68320f4-4819-4b89-99ff-8dbd0c1e1518?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D24125419-f68320f4-4819-4b89-99ff-8dbd0c1e1518%26entityType%3Dcollection%26workspaceId%3D66e2058d-1f10-4f8a-8dd0-5aac8f603740)

## Разработанный функционал
1. модуль авторизации
2. подключена бд Postgres
3. реализован механизм подписок на пользователей
4. реализован функционал отправления уведомления на почту
5. реализован механизм объединения нескольких уведомлений в одно письмо

Пример уведомления отправленного пользователю
```
Сегодня 14-06 день рождения празднуют:
1) test@yandex.test
2) test2@yandex.test


не забудь поздравить своих коллег!
```
