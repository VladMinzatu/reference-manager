# reference-manager

## Starred references

The application and db schema now include the possibility to mark certain references as 'starred', along with the possibility to filter out references when retrieving them based on their starred status.

My personal use for this functionality is to highlight certain reference books or sources (usually books) that are key references. Usually, these will have the following characteristics:
- they usually cover a broad topic (e.g. operating systems, microservices, etc.)
- they are exceptionally clear, providing intuitive explanations as well as going into sufficient depth in the topics covered
- they give a great overview of the whole topic, as well as pointing to references for digging deeper

Some of my favorite examples of such references would be:
- [Code: The Hidden Language of Computer Hardware and Software](https://isbnsearch.org/isbn/9780137909100) by Charles Petzold
- [Designing Data Intensive Applications](https://isbnsearch.org/isbn/9781449373320) by Martin Kleppmann
- [Operating Systems: Three Easy Pieces](https://isbnsearch.org/isbn/9781985086593) by Remzi H Arpaci-Dusseau; Andrea C Arpaci-Dusseau
- [Learning Domain-Driven Design: Aligning Software Architecture and Business Strategy](https://isbnsearch.org/isbn/9781098100131) by Vlad Khononov
- [Hypermedia Systems](https://isbnsearch.org/isbn/9798394025143) by Carson Gross, Adam Stepinski, Deniz Akşimşek
- [Learning Go: An Idiomatic Approach to Real-World Go Programming](https://isbnsearch.org/isbn/9781098139292) by Jon Bodner

and others to be found in the db.


## Domain and DB modeling notes

The main assumption behind the modeling choices made here is that the number of categories, as well as the number of references within one category are relatively small (few 100s at most).

That said, the domain model and thus db schema are constructed in such a way that transactional boundaries are as granular as possible. It may have been feasible to treat whole categories (with all their references) as one aggregate, but what would be the fun in that? :) And of course, the more granular boundaries allow us to work with more efficient (lighter) requests and end users tend to like those.

The exercise here is to model the data storage and queries in a way that data integrity and performance are guaranteed in the case of highly concurrent usage. Think of this schema being used in a situation where this data is joined with a user_id coming from somewhere, where each user has their own list of categories with the references within, although that part is not modeled here (for now).

### Notes on ordering

One interesting choice when it comes to the db schema was to store the positions of the categories or references in the same table as the id.

Separate tables may seem like a cleaner approach, and indeed, other options were explored first (check the commit history), but the need to maintain consistency and handle concurrency issues with positions while performing mutating operations did make this the cleaner, easier to manage and more performant choice.

### Notes on concurrency and locking

We need to ensure data integrity of the positional data on concurrent writes (adding new categories or adding new references within a category).

The `SELECT FOR UPDATE` pattern can be used to achieve row-level locking during a transaction, if we were using Postgres for example. e.g. `SELECT COALESCE(MAX(position) + 1, 0) FROM reference_positions WHERE category_id = $1 FOR UPDATE` and then insert at that position (rows for that cateogry are locked until the transaction is committed or rolled back), which would work fine.

But Sqlite doesn't support `SELECT FOR UPDATE`, so we have organised our queries in a way that would not be safe against race conditions in Postgres at times, taking advantage of SQLite's db-locking semantics. So we are safe, given that we are using SQLite, but if we were to switch to Postgres, the queries and code in our `adapters/repository.go` would need to suffer some changes, because of different concurrency guarantees. Comments are left in the code around most queries where we take a simple approach that is afforded by SQLite, but where we'd have to do some explicit locking if we were to use e.g. Postgres.

### Side note on modeling

Since we're talking about alternative DBs, fine grained transactions can be simulated with coarser underlying aggregates (think a document db with document per category), where updates are applied to the aggregate using optimistic locking. (generally with a tradeoff in the number of conflicts, but for what we're doing here, would be a decent choice and would simplify some things).
So in an alternative implementation, we could keep the exact same interface as here and use an underlying document store with one entry per category (with all the references data within that entry). And all the modeling (and schema evolution and concurrency) implications that that brings. Or we could change the interface to reflect such coarse grained logic as well. But that's not the approach we took.

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
