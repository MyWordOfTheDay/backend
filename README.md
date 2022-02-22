# Backend

Provides a gRPC and optional HTTP backend for MyWordOfTheDay

# Pre-requisites

- Install [grpcurl](https://github.com/fullstorydev/grpcurl)
- Go >= 1.16. Because we use embed, any versions older than 1.16 won't work.

# gRPC requests

## List services

```
grpcurl -plaintext localhost:8080 list
```

## List RPC endpoints

Note: This assumes the api repository is cloned at the same location as this repository (`../api`), that the required dependencies have been install into the `.cache` directory, of the api repository, and that you're using an Apple device (Darwin).

```
grpcurl -protoset <(cd ../proto; ../proto/.cache/Darwin/x86_64/bin/buf image build -o -) -plaintext localhost:8080 list mywordoftheday.v1alpha1.MyWordOfTheDayService
```

# HTTP requests

## Add Word

```
curl -H "Content-Type: application/json" -X POST localhost:8443/api/v1alpha1/word -d '{"word": "floccinaucinihilipilification"}'
```

## List Words

```
curl -H "Content-Type: application/json" -X GET localhost:8443/api/v1alpha1/words
```

## Delete Word

```
curl -H "Content-Type: application/json" -X DELETE localhost:8443/api/v1alpha1/word/1
```

# Running the Dockerfile

## Build the image

```
docker build -f ./Dockerfile -t mywordoftheday .
```

## Run the image

```
docker run -p 8443:8443 -v /path/to/config:/config mywordoftheday
```

## Publish the docker image

```
docker login

docker tag mywordoftheday simondrake/mywordoftheday:v1alpha1

docker push simondrake/mywordoftheday:v1alpha1
```

# Tests

Tests are separated into two separate categories

## Unit Tests

Unit Tests do not require any set-up steps and can simply be run with `make test`

## Integration Tests

Integration Tests require actual services to be up (e.g. Postgres), so require a bit of set-up before they can be run.

* Run the test database (`docker-compose -f docker-compose-integration-tests.yaml up -d`)
* Run `make integration-test`

**Note:** If you need to login to the db container, you can do so with the following command:

```
docker exec -it integration_tests_db psql postgresql://user:password@localhost:5432/mywordofthedaytests
```
