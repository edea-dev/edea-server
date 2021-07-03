# EDeA Backend

This contains the program `edead` which handles the API, serving the portal web assets and caching of modules.

## Running it

### Grab the latest tagged release:
 from [here](-/tags) (to be done), unpack it, adjust the database connection information in `config.yml` and run it like this:

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

0. Install [modd](https://github.com/cortesi/modd) for live reloading and auto-building

```sh
env GO111MODULE=on go get github.com/cortesi/modd/cmd/modd
```

Alternatively, use your system's package manager. On Arch Linux, if you use `yay`, just run `yay -S modd` 

1. Clone the repository

```sh
git clone https://gitlab.com/edea-dev/edead && cd edead
```

2. Configure the database connection

Copy `config.template.yml` to `config.yml`, and edit it to point to your PostgreSQL installation. Installing and configuring PostgreSQL is left as an exercise for the reader.

3. Run it

```sh
modd
```

That's it.

## Deployment

```sh
rsync -avz static edea.dev:~/edea-test/
scp -C edead edea.dev:~/edea-test/
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
