```cmd
migrate create -seq -ext .sql -dir ./migrations create_movies_table
migrate -path ./migrations -database "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" up
migrate -path ./migrations -database "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" version
migrate -path ./migrations -database "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" goto 1
migrate -path ./migrations -database "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" down 1
migrate -path ./migrations -database "postgres://greenlight:greenlight@localhost/greenlight?sslmode=disable" force 1

```

https://github.com/golang-migrate/migrate/issues/826

```cmd
CREATE DATABASE greenlight;

CREATE ROLE greenlight WITH LOGIN PASSWORD 'greenlight';

grant all privileges on database greenlight to greenlight;

alter database greenlight owner to greenlight;

CREATE EXTENSION IF NOT EXISTS citext;
```

race condition
```cmd
seq 1 10 | xargs -I % -P8 curl -X PATCH -d "{\"runtime\": \"97 mins\"}" "localhost:4000/v1/movies/4"

seq 1 6 | xargs -I % -P8 curl "http://localhost:4000/v1/healthcheck"

curl -w "\nTime: %{time_total}s \n" localhost:4000/v1/movies/1

curl -d "{\"name\": \"Edith Smith\", \"email\": \"edith2@example.com\", \"password\": \"pa55word\"}" localhost:4000/v1/users & windows-kill -SIGINT 8016

netstat -a -o
taskkill /F /PID pid_number

curl localhost:4000/v1/healthcheck & taskkill /F /PID 15024

hey -d "{\"email\": \"alice@example.com\", \"password\": \"pa55word\"}" -m "POST" http://localhost:4000/v1/tokens/authentication
```

winget install GnuWin32.Make
add env path C:\Program Files (x86)\GnuWin32\bin
restart terminal

go tool dist list

go env GOCACHE

go build -a -o=/bin/foo ./cmd/foo        # Force all packages to be rebuilt
go clean -cache                          # Remove everything from the build cache