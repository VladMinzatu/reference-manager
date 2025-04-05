# reference-manager

## Domain and DB modeling notes

The main assumption behind the modeling choices made here is that the number of categories, as well as the number of references within one category are relatively small (few 100s at most).

That said, the domain model and thus db schema are constructed in such a way that transactional boundaries are more as granular as possible. It may have been feasible to treat whole categories as one aggregate, but what would be the fun in that? :) And of course, the more granular boundaries lead to more efficient (lighter) requests and end users tend to like those.

The exercise here is to model the data storage and queries in a way that data integrity and performance are guaranteed in the case of highly concurrent usage. Think of this schema being used in a situation where this data is joined with a user_id coming from somewhere, where each user has their own list of categories with the references within, although that part is not modeled here (for now).

## Working with the db locally

We use `slite3` as our db and [goose](https://github.com/pressly/goose) to manage migrations.

Follow the instructions on the goose page above to install the tool and add it to your path.

Next, let's look at some useful instructions:

Initialising the db by running migrations:

```
goose sqlite3 db/references.db -dir db/migrations up
```

Connecting to our db to run queries:

```
sqlite3 db/references.db

sqlite> .tables
```

Tearing down the db:

```
goose sqlite3 db/references.db -dir db/migrations down
```
