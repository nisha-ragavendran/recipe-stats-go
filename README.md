# Recipe Stats

CLI tool that reads a JSON file of recipe delivery records and outputs stats.

## Run locally

```bash
go run . --file {filepath}
```

With optional flags
```bash
go run . --file {filepath} --postcode 10120 --from 10AM --to 3PM --recipe Chicken
```

## Run with Docker

```bash
docker build -t recipe-stats-go -f Dockerfile .

docker run --rm -v $(pwd)/demo_100.json:/data/recipes.json \
  recipe-stats --file /data/recipes.json --postcode 10120 --from 10AM --to 3PM --recipe Chicken
```

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--file` | yes | path to JSON input file |
| `--postcode` | no | postcode for time-window delivery count |
| `--from` | no | start of time window e.g. `10AM` |
| `--to` | no | end of time window e.g. `3PM` |
| `--recipe` | no | partial recipe name to search |