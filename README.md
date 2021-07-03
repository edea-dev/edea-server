# EDeA Backend

This contains the program `edead` which handles the API, serving the portal web assets and caching of modules.

## Running it

### Grab the latest tagged release:
 from [here](#) (to be done), unpack it, adjust the database connection information and run it like this:

```sh
./edead
```

This should output helpful log information to the standard output in case something goes wrong.


### Or if you have a working Go installation, you can just run:

```sh
go get gitlab.com/edea-dev/edead/cmd/edead
go install gitlab.com/edea-dev/edead/cmd/edead
$GOPATH:/bin/edead
```

## Development

0. Install [modd](https://github.com/cortesi/modd) for live reloading (optional)

```sh
env GO111MODULE=on go get github.com/cortesi/modd/cmd/modd
```

1. Clone the repository

```sh
git clone https://gitlab.com/edea-dev/edea
```

2. Run it

```sh
cd edea/backend/
go build gitlab.com/edea-dev/edead/cmd/edead
./edead

# or with modd for live code reloading:
$GOPATH/bin/modd
```

That's it.

## Deployment

```sh
rsync -avz static edea.dev:~/edea-test/
scp -C edead edea.dev:~/edea-test/
```

## Administration

We'll add any routine administrative tasks to the documentation as they arise after the portal goes live on [edea.dev](https://edea.dev).

## Assorted Tasks

- GraphQL supported hosters:
  - GitHub API client
  - Gitea API client
  - sr.ht API client? maybe?

- web hooks for hosters
  - tie to login

- Task runners
  - update repos from external

- Caching of arbitrary data
- Fetch repositories
  - render schematic and layout files
    - cache them
