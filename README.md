# EDeA Backend

This contains the program `edead` which handles the API, serving the portal web assets and caching of modules.

## Running it

### Grab the latest tagged release

 from [here](-/tags) (to be done), unpack it, adjust the database connection information in `config.yml` and run it like this:

```sh
./edead
```

This should output helpful log information to the standard output in case something goes wrong.

### Or if you have a working Go installation, you can just run

```sh
go get gitlab.com/edea-dev/edead/cmd/edead
go install gitlab.com/edea-dev/edead/cmd/edead
$GOPATH:/bin/edead
```

## Development

0. Install Dependencies

We use [tilt](https://tilt.dev/) for live updating and docker compose for the database and search images.

1. Clone the repository

```sh
git clone https://gitlab.com/edea-dev/edead && cd edead
```

2. Run a local dev environment with tilt

```sh
tilt up
```

That's it.

### Running tests

To run the frontend tests:

```sh
cd frontend
npx playwright test
```

## Administration

We'll add any routine administrative tasks to the documentation as they arise after the portal goes live on [edea.dev](https://edea.dev).

## Dependencies

edead needs a variety of tools to be available on the host to function:

- python >=3.8
  - merge tool, plotpcb
- svgcleaner
  - svg preparation for caching
- KiCAD 5.x
  - plotpcb
- plotgitsch
  - Plotting differences in the schematic
- mdbook
  - rendering of the documentation
