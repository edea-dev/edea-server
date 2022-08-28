# How to set up your own instance

This assumes building the edea-server from source and using a hardcoded set of credentials in the configuration.
For using other authentication providers, anything supporting OpenID Connect should work, please refer to the documentation of the respective provider for that.

## Prerequisites

Most distributions should ship those packages as they're very common, for MacOS or Windows the website
of the respective tools should also provide information.

- Go compiler (<https://go.dev/dl/>)
- yarn (<https://yarnpkg.com/>)
- unix tools such as bash and sed (for Windows Cygwin or MinGW should provide those)
- python 3.10 (<https://www.python.org/>) or newer
- PostgreSQL database
- edea tool (<https://gitlab.com/edea-dev/edea>)
- Optional: Meilisearch (<https://www.meilisearch.com/>) for fulltext search

## Building

Run the following steps to produce `edea-server` binary and the `static` assets folder:

```sh
git clone https://gitlab.com/edea-dev/edea-server.git
cd edea-server
mkdir static # for the js, css, etc. files
make deps # build the frontend files
make build # build the executable
```

For running the portal you only need the server binary, the static folder with the assets and a configuration file.

## Configuration

Let's go over [the configuration](https://gitlab.com/edea-dev/edea-server/-/blob/main/config.template.yml) step by step:

```yaml
server:
  host: localhost
  port: 3000
```

Host and port are of course where the instance should run on. If you want to run it on a public facing web-server `0.0.0.0` for the host and `80` for the port would be advised, there's no need for a frontend proxy server.

```yaml
dsn: host=127.0.0.1 user=edea password=edea dbname=edea port=5432 sslmode=disable
```

The `dsn` string is for the database connection. It needs a PostgreSQL database for now, but we might implement SQLite support in the future. Host, user, password and dbname are of your database, port is the default one for PostgreSQL and `sslmode=disabled` is there because in this case we're running both on the same host so no need to encrypt the database connection.

```yaml
cache:
  repo:
    base: /home/user/git/edea/backend/tmp/git
  book:
    base: /home/user/git/edea/backend/tmp/doc
  diff:
    schema: ./tmp/sch
    layout: ./tmp/pcb
```

Cache folders. Those paths specify where the repositories, the documentation for modules and the schema and layout diffing files can be stored. The `diff` folders hold only temporary files but the `repo` and `book` folders cache files which are used constantly.

```yaml
auth:
  oidc:
    provider_url: http://your-hostname:3000
    client_id: a-random-id
    client_secret: the-client-secret
    redirect_url: http://your-hostname:3000/callback
    logout_url: http://your-hostname:3000/logout_callback
    post_logout_url: http://your-hostname:3000/
  oidc_server: 
    use_builtin: true
    post_logout_urls:
      - http://your-hostname:3000/logout_callback
    redirect_urls:
      - http://your-hostname:3000/callback
```

Now we get to the authentication. edea-server includes it's own OIDC compatible, but not compliant (it's only the bare minimum) authentication server. This is good enough for just a few users you trust but wholly inadequate for hosting many users. `users.yml` contains the test users, more can simply be added. To generate the password hashes run `echo -n "password" | argon2 testsalt -id -e`. The salt is embedded in the output so you can also generate a random salt for each user for better security.

There's hosted and open source solutions like [Ory](https://www.ory.sh) which have free plans for testing and open source projects but are also self-hostable. The OpenID website has a [list of certified solutions](https://openid.net/developers/certified/) that should also work.

You can specify many providers with different keys if you want, but only the one with the name `oidc` will be used. In the above example it's configured using the builtin provider though.
If `use_builtin: true` is set, it will run the builtin provider when `edea-server` starts. You can also specify more URLs for `redirect_urls` and `post_logout_urls` if you want to use the same config for testing and production.
Just make sure that the URLs under `oidc` are the correct ones for your currently running instance.

Make sure that you're actually setting the hostname and not an IP address for the various URLs though as it won't work otherwise.

```yaml
search:
  host: http://meili-host:7700
  index: edea
  api_key: meiliedea
```

Finally, this brings us to the Meilisearch settings. While the feature is optional, the configuration section is not, so don't forget to add it even if the values are empty or invalid.

It just needs the search host, the index name and an `api_key` with the permissions `search`, `documents.add`, `documents.get`, `documents.delete`, `tasks.get` and `version`.
To set this up via curl:

```sh
# create the index
curl \
  -X POST 'http://localhost:7700/keys' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer MASTER_KEY' \
  --data-binary '{
    "uid": "edea",
    "primaryKey": "id"
  }'

# create an api key with the right permissions
curl \
  -X POST 'http://localhost:7700/keys' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer MASTER_KEY' \
  --data-binary '{
    "description": "edea index key",
    "actions": ["search", "documents.add", "documents.get", "documents.delete", "tasks.get", "version"],
    "indexes": ["edea"],
    "expiresAt": "2030-01-01T00:00:00Z"
  }'

# make the index filterable by the user_id and public attributes
curl \
  -X PATCH 'http://localhost:7700/indexes/edea/settings' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer MASTER_KEY' \
  --data-binary '{
  "filterableAttributes": ["user_id", "public"]
  }'
```

## Installing the edea tool

Before actually starting the server, the edea tool also needs to be available.
Installing it is as simple as running `pip install edea` (in case at time of reading v0.1 has been released on pypi.org).

In case it is not yet available or if you want a development version, follow those steps:

```sh
git clone https://gitlab.com/edea-dev/edea
cd edea
poetry build
pip install edea-0.1.0-py3-none-any.whl
```

The resulting `.whl` file can also be copied to a server and installed there, or it can be installed in a virtual environment for a user-only install.

## Running it

Now that the configuration file is written to `config.yml` you can just run edea-server and start tinkering with it. The log output will be displayed on the console.

## Meilisearch

As Meilisearch has still not reached 1.0 it's best to check the official [Quickstart section](https://docs.meilisearch.com/learn/getting_started/quick_start.html) of the documentation if you want to run your own server.
