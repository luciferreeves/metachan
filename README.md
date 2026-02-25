> [!CAUTION]
> I _do not_ provide pre-hosted instances of the MetaChan API. You will need to host your own instance of the API to use it. MetaChan is provided on an "as-is" basis and fetches data from various public sources and projects, some of which may have rate limits, restrictions, and even contain pirated content. The API is intended for personal use only and should not be used for commercial purposes. I do not take any responsibility for any legal issues that may arise from using the API. Please use it at your own risk. I _do not_ condone piracy or illegal distribution of content.

# MetaChan

Welcome to **MetaChan**. MetaChan is an Anime and Manga metadata API that provides a RESTful interface for accessing metadata for various anime and manga titles. MetaChan best integrates with [MyAnimeList](https://myanimelist.net/) and uses **MAL IDs** as the primary identifier for anime and manga titles. [AniList](https://anilist.co/) is also supported partially and will reverse lookup MAL IDs.

<div align="center">
<img src="https://i.redd.it/rg4mpacfm1wz.png" width="730">

[![](https://tokei.rs/b1/github/luciferreeves/metachan?category=code&style=for-the-badge)](https://github.com/luciferreeves/metachan) [![](https://tokei.rs/b1/github/luciferreeves/metachan?showLanguage=true&languageRank=1&label=Top%20Language&style=for-the-badge)](https://github.com/luciferreeves/metachan) ![GitHub License](https://img.shields.io/github/license/luciferreeves/metachan?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/luciferreeves/metachan?style=for-the-badge) ![GitHub Repo stars](https://img.shields.io/github/stars/luciferreeves/metachan?style=for-the-badge)

</div>

## Prerequisites

- [Go](https://go.dev) 1.24.1 or later
- [Make](https://www.gnu.org/software/make/)

If you wish to use Docker, you will also need to install [Docker](https://www.docker.com). Additionally, you may need to install one of the supported databases if you wish to use a database other than SQLite.

- [PostgreSQL](https://www.postgresql.org/)
- [MySQL](https://www.mysql.com/)
- [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server/sql-server-downloads)
- [SQLite](https://www.sqlite.org/index.html) (not required, but recommended for development)

You might also need to install [Git](https://git-scm.com/) if you want to clone the repository.

## Configuring Environment Variables

You can configure the environment variables by creating a `.env` file in the root directory of the project. The example `.env` file contains all the available environment variables and their default values. You can modify these values to suit your needs.

The following environment variables are available:

| Variable Name | Description | Default Value |
| --- | --- | --- |
| `HOST` | The host address to bind the server to. | `0.0.0.0` |
| `PORT` | The port to run the API on. | `3000` |
| `DEBUG` | Enable debug logging. | `false` |
| `DB_DRIVER` | The database driver to use. Supported drivers are `sqlite`, `postgres`, `mysql`, and `sqlserver`. The options are **case-sensitive**. | `sqlite` |
| `DSN` | The Data Source Name (DSN) for the database connection. The format depends on the database driver you are using. See [Configuring Data Source Names (DSN)](#configuring-data-source-names-dsn) below. | `metachan.db` |
| `ANISYNC` | Enable background anime sync task. | `false` |
| `TMDB_API_KEY` | API key for [TMDB](https://www.themoviedb.org/) episode enrichment. | |
| `TMDB_READ_ACCESS_TOKEN` | Read access token for TMDB API v4. | |
| `TVDB_API_KEY` | API key for [TVDB](https://thetvdb.com/) episode enrichment. | |

### Configuring Data Source Names (DSN)

The DSN format varies depending on which database driver you are using. Refer to the table below for the correct format for each driver.

| Driver | DSN Format |
| --- | --- |
| `sqlite` | `metachan.db` or `/path/to/metachan.db` |
| `postgres` | `host=localhost port=5432 user=postgres password=postgres dbname=metachan sslmode=disable` |
| `mysql` | `root:password@tcp(localhost:3306)/metachan` |
| `sqlserver` | `sqlserver://localhost:1433?database=metachan` |

## Local Development

After you have installed the prerequisites, [forked](https://github.com/luciferreeves/metachan/fork) the repository and cloned it to your local machine, start by installing the dependencies:

```bash
make setup
```

This will create an example `.env` file in the root directory of the project. You can use this file to configure the environment variables for your local development environment and database connection.

Now you can start the development server by running:

```bash
make dev
```

## Building for Production

> [!WARNING]
> The API is still under **heavy development** and the `main` branch contains breaking changes. A lot of features are still missing and the Documentation is not complete. There are _no releases_ yet. If you still want to use the API, you can build it from the source code or use the [Dockerfile](Dockerfile) to build a Docker image. The API is not production ready yet and should be used at your own risk. I am not responsible for any data loss or damage caused by using the API.

To build for production, run the following commands:

```bash
make build
make run
```

This will build the binary and run it. The API will be available at `http://localhost:3000` by default.

## Docker

You can also run the API using Docker. To build the Docker image, run the following command:

```bash
docker build -t metachan .
```

To run the Docker image:

```bash
docker run -p 3000:3000 -e DB_DRIVER=sqlite -e DSN=metachan.db --name metachan metachan
```

This will expose the API on port `3000`. You can change the `DB_DRIVER` and `DSN` environment variables to use a different database.

## API Documentation

> [!NOTE]
> Documentation for this API is **not complete**. The API is still under heavy development and the endpoints are subject to change. This [README](README.md) file will be updated with the API documentation once it is available.