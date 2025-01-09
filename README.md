# Go-Monko

## Setup
Copy `secrets.json.example` to `secrets.json` and fill in with the relevant information. The Discord Bot must have permission to send links.

`Limit` is the amount of tokens that the transaction must be over to send to discord.  This allows you to filter out smaller transactions

## Running the project

You can use Docker Compose to run the project:

```sh
docker-compose up -d --build
```

If you have a local installation of Go, you can run with:
```sh
go run watch.go
```