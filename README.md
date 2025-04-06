# reference-manager

## Domain and DB modeling notes

The main assumption behind the modeling choices made here is that the number of categories, as well as the number of references within one category are relatively small (few 100s at most).

That said, the domain model and thus db schema are constructed in such a way that transactional boundaries are as granular as possible. It may have been feasible to treat whole categories as one aggregate, but what would be the fun in that? :) And of course, the more granular boundaries allow us to work with more efficient (lighter) requests and end users tend to like those.

The exercise here is to model the data storage and queries in a way that data integrity and performance are guaranteed in the case of highly concurrent usage. Think of this schema being used in a situation where this data is joined with a user_id coming from somewhere, where each user has their own list of categories with the references within, although that part is not modeled here (for now).

### Row-level locking when inserting a new category or reference

We need to ensure data integrity of the positional data on concurrent writes (adding new categories or adding new references within a category).

The `SELECT FOR UPDATE` pattern can be used to achieve row-level locking during a transaction, if we were using Postgres for example. e.g. `SELECT COALESCE(MAX(position) + 1, 0) FROM reference_positions WHERE category_id = $1 FOR UPDATE` and then insert at that position (rows for that cateogry are locked until the transaction is committed or rolled back), which would work fine.

But Sqlite doesn't support `SELECT FOR UPDATE`, so we need to use a different pattern (that will work with Postgres as well): create a position_sequence table and `INSERT...ON CONFLICT DO UPDATE...RETURNING`. In Postgres, this will lock that row, allowing us to achieve what we want. In Sqlite, it will also work, but via different locking mechanics: whole database (in default mode) or WAL segment (in WAL mode).

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

Creating a new migration:

```
goose create add_position_sequence_tables sql -dir ./db/migrations
```
