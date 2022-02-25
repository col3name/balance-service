### Requirements
[make](https://www.gnu.org/software/make/), [docker](https://www.docker.com/), [docker-compose](https://docs.docker.com/compose/install/)


#### optional
[golangci-lint](https://github.com/golangci/golangci-lint),
[apache benchmark](https://httpd.apache.org/docs/2.4/programs/ab.html)

setup data

insert data from file ```data\postgres\migrations\payment_public_account.sql```
```shell
make bin/money
make up
```

Ограничения на количество запросов к api для получения курса валют.

Поддерживаю конвертацию в EUR, USD

Пагинация с помощью курсора