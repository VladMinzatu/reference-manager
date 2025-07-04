# reference-manager

> **⚠️ Note**
>
> With this side project, I mainly focused on the domain model and internal architecture (i.e. everything under the `/domain` and `/adapters` directories). This has been more carefully considered and is also described in the [Domain and DB modeling notes](#domain-and-db-modeling-notes) section below.
>
> As of now, other parts of the project - in particular, the CLI (`cmd`) and `web` interfaces - haven't received the same level of attention and were rather more rushed and probably contain areas of suboptimal design and performance.
>
> Also, many production aspects like logging, observability or proper configuration handling haven't been added, as this is not the focus for me for this project.

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

In this app, we have references grouped into categories and both the list of categories, as well as the list of references within a given category have to maintain a specific ordering. That's it.

As simple as this app is, there are lots of competing choices and tradeoffs to be discussed in terms of domain modeling, and that's what I will cover in this section. But first, let's identify the assumptions and requirements:

### Assumptions

- **A1**: The number of categories, as well as the number of references within one category are relatively small (few 100s at most).
- **A2**: The application will need to support highly concurrent usage (I mean, not really, but this is the exercise I'm setting up for myself here :) ). But not the kind where the same category or reference is accessed concurrently (except maybe rarely). Think of it more as the schema being used in a situation where this data is joined with a user_id coming from somewhere, where each user has their own list of categories with the references within. Although that part is not modeled here. The point is, we'll need performant queries and granular locking.

### Requirements

- **R1**: Transactional integrity of the data must be guaranteed: from ordering of categories and references to relations between categories and their references, the data must be strongly consistent and not eventually consistent. So transactional boundaries are very important.
- **R2**: Performance and granularity: We should minimise data transfer between the app and DB as far as reasonably possible and avoid heavy queries: in particular, we cannot accept a solution where updating a references or adding a new references requires rewriting the entire data of the whole category to the db each time.

### Why not a transaction script pattern?

What we are seeing so far from those assumptions and requirements is the need for performance and consistency. It also looks like we don't have domain objects that are behavior rich or complex business logic (we have data-centric logic). It seems like a transaction script pattern would be a good fit here.

The main argument against it is if we want to future-proof the code for the application logic growing in the future. And although this is just a side project, I will assume this future proofing is important, just like it would be in a real product development scenario.

So we'll try to build a proper DDD domain model for our app.

### The domain model

In attempting to build a nice clean DDD-approved domain model, we immediately hit a wall that emerges out of our requirements for both performance and strong consistency.

First of all, since for the sake of performance we don't want to save an entire category's contents when adding or updating a reference, we cannot have references be part of a category aggregate that defines the transaction boundary and protects it via optimistic locking in the classic DDD sense. This suggests that we might have to model our aggregates more granularly: categories are aggregates and references are aggregates and they refer to each other through ids.

But this leads us to a big DDD no-no: because we said we wouldn't tolerate eventual consistency, now we need to enforce invariants and maintain transactional consistency across aggregates to maintain the ordering of both references and categories consistent. Is that even possible? Who is going to protect that transactional boundary?

It's obvious we're going to have to get creative and break some rules. We should just be wise about which ones we choose to break. Here's what I'm thinking of doing:

- Use the Category as the single aggregate in our domain model. That way, the category protects the consistency of the references data, including ordering, by providing the transactional boundary through optimistic locking. So far so good.
- But what about the performance of adding/deleting a new reference or editing an existing reference or even reordering? We don't want to write the whole aggregate to the db each time. For this, we will have a dedicated ReferencesRepository to perform the operations at a reference level without involving the aggregate. This breaks the rule that entities should only be manipulated via methods in their aggregates. But we have to be pragmatic, and the fact that we sacrifice neither performance nor consistency is worth it.
- And how do we manage the ordering of the categories? Who provides the locking/transactional boundary there? This is now a multi-aggregate operation. This is achieved through a dedicated service+repository that uses DB row-level locking to enforce consistency at the service level, outside any aggregate. Note that we could have done this for references in a category as well and made both categories and references their own aggregates. But since this is a side project, I'm welcoming the chance to skin a cat multiple ways.
- We can also add unique compound index constraints on (category_id, position) and (reference_id, position) at the DB level for some infrastructure support with maintaining data consistency.

### Was it even worth looking at this through the lens of DDD in the first place?

TBD

### Notes on ordering

One interesting choice when it comes to the db schema was to store the positions of the categories or references in the same table as the id.

Separate tables may seem like a cleaner approach, and indeed, other options were explored first (check the commit history), but the need to maintain consistency and handle concurrency issues with positions while performing mutating operations did make this the cleaner, easier to manage and more performant choice.

### Notes on concurrency and locking

We need to ensure data integrity of the positional data on concurrent writes (adding new categories or adding new references within a category).

The `SELECT FOR UPDATE` pattern can be used to achieve row-level locking during a transaction, if we were using Postgres for example. e.g. `SELECT COALESCE(MAX(position) + 1, 0) FROM reference_positions WHERE category_id = $1 FOR UPDATE` and then insert at that position (rows for that cateogry are locked until the transaction is committed or rolled back), which would work fine.

But Sqlite doesn't support `SELECT FOR UPDATE`, so we have organised our queries in a way that would not be safe against race conditions in Postgres at times, taking advantage of SQLite's db-locking semantics. So we are safe, given that we are using SQLite, but if we were to switch to Postgres, the queries and code in our `adapters/repository.go` would need to suffer some changes, because of different concurrency guarantees. Comments are left in the code around most queries where we take a simple approach that is afforded by SQLite, but where we'd have to do some explicit locking if we were to use e.g. Postgres.

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
