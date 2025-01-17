# Monko Watcher

## Setup
Copy `config.yaml.example` to `config.yaml` and fill in with the relevant information. The Discord Bot must have permission to send links.

## Running the project

You can use Docker Compose to run the project:

```sh
docker-compose up -d --build
```

If you have a local installation of Go, you can run with:
```sh
go run watch.go
```
