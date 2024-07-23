# Go Final Project

Этот проект представляет собой веб-сервер на Go для управления задачами с использованием базы данных SQLite. Веб-сервер поддерживает создание, редактирование, удаление и выполнение задач, а также предоставляет возможность аутентификации с использованием JWT.

## Описание проекта

Этот веб-сервер был создан для управления задачами. Он позволяет добавлять, редактировать, удалять и отмечать задачи как выполненные. Пользователи могут также искать задачи по заголовку, комментариям или по дате. Для доступа к серверу требуется аутентификация по паролю.

## Выполненные задания со звёздочкой

- Реализована возможность определять извне порт при запуске сервера. Если существует переменная окружения `TODO_PORT`, сервер при старте слушает порт со значением этой переменной.
- Реализована возможность определять путь к файлу базы данных через переменную окружения. Сервер получает значение переменной окружения `TODO_DBFILE` и использует его в качестве пути к базе данных, если это не пустая строка.
- Добавлена возможность выбора задач через строку поиска. Проверяется наличие строки поиска в заголовке или комментарии задач. Также можно выбрать задачи на конкретную дату, проверяя формат `02.01.2006`.
- Реализована аутентификация пользователей с использованием JWT. Доступ к задачам возможен только после успешной аутентификации.

## Инструкция по запуску кода локально

1. Клонируйте репозиторий:
    ```sh
    git clone https://github.com/yourusername/go_final_project.git
    cd go_final_project
    ```

2. Установите необходимые зависимости:
    ```sh
    go mod download
    ```

3. Создайте файл `.env` и укажите необходимые переменные окружения:
    ```env
    TODO_PORT=7540
    TODO_DBFILE=./scheduler.db
    TODO_PASSWORD=mypassword
    ```

4. Запустите сервер:
    ```sh
    go run main.go
    ```

5. Откройте браузер и перейдите по адресу `http://localhost:7540`.

## Инструкция по запуску тестов

1. Убедитесь, что переменная `Token` в файле `tests/settings.go` содержит токен, полученный от сервера.
2. Запустите тесты:
    ```sh
    go test ./tests
    ```

## Инструкция по сборке и запуску проекта через Docker

1. Соберите Docker-образ:
    ```sh
    docker build -t go_final_project .
    ```

2. Запустите контейнер:
    ```sh
    docker run -d -p 7540:7540 \
        -e TODO_PASSWORD=mypassword \
        -e TODO_DBFILE=/app/scheduler.db \
        -v /path/to/your/host/db:/app/scheduler.db \
        --name go_final_project_container \
        go_final_project
    ```

    Замените `/path/to/your/host/db` на реальный путь к вашей базе данных на хост-машине.

3. Откройте браузер и перейдите по адресу `http://localhost:7540`.

## Заключение

Этот проект реализует веб-сервер для управления задачами с использованием базы данных SQLite и аутентификации с использованием JWT. Все задания, включая задания со звёздочкой, были успешно выполнены.
