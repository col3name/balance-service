# Service for working with user balance
![Ci passing](https://github.com/col3name/balance-service/actions/workflows/prod.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-0.0%25-red)

## TODO
- Write test
- Setup Github CI

## Requirements
[make](https://www.gnu.org/software/make/), [docker](https://www.docker.com/), [docker-compose](https://docs.docker.com/compose/install/)

### optional
[golangci-lint](https://github.com/golangci/golangci-lint),
[apache benchmark](https://httpd.apache.org/docs/2.4/programs/ab.html),
[newman](https://www.npmjs.com/package/newman),
[go-cleanarch](https://github.com/roblaszczak/go-cleanarch)

k6 ```brew instal k6```


### Initialize database
create table from file ```data\postgres\migrations\init.sql```

insert data from file ```data\postgres\migrations\payment_public_account.sql```

### Start  app
```shell
make bin/money
make up
```

### Benchmark for read transactions 
```shell
make bench
```

server running on [localhost:8000](http://localhost:8000) ```http://localhost:8000```

swagger ui running on [localhost:80](http://localhost:80) ```http://localhost:80```

## Use cases 
- Get the user's current balance.
- Get the transaction history of the account.
- Transfer funds from user to user.
- Debit or credit of funds from account.
- Converting balance to other currencies.

## Improvments
 - Use Credit/Debit transaction account, Transaction(id, account_id uuid, credit int, debit int, desc: string)
## Database UML

![cursor](docs/img/db-uml.png)

### Pagination when sorting by date is implemented via the cursor to optimize the OFFSET operation

The user rarely looks at the entire transaction history when sorting by any gender.
Therefore, it is possible to save data for users and update it during financial transactions.


![cursor](docs/img/pagination-compare.png)
![cursor](docs/img/pagination-compare-2.png)

### Improvements in architecture to increase fault tolerance

![cursor](docs/img/sequence.png)
![cursor](docs/img/components.png)
