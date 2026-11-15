# Spike: MongoDB for AREA persistence

## Goal

The other persistence spikes use relational databases (PostgreSQL, SQLite). This one tests a **document database** on four things AREA actually needs, using the Python driver of the FastAPI spike:

1. Store a whole workflow (trigger, condition, reactions) and read it back.
2. Reject the same hook event twice (the subject requires hooks not to treat the same occurrence as new indefinitely).
3. Refuse a workflow that references a user that does not exist.
4. Run a multi-document transaction (e.g. record a run and mark the event as processed atomically).

Prior experience: never used MongoDB before.

## Timebox

**45 minutes**, fixed before starting (20:52). Shorter than the backend spike because the server code already existed; only the database and the script were new.

## Where I was when the time ran out

Done at 20:53, after about **1 minute** of execution (script written with AI assistance, then run and reviewed by me). All four checks ran and gave a clear answer:

| # | Check | Result |
|---|---|---|
| 1 | Workflow as one document | **Works.** The nested trigger / condition / two branches with 4 reactions fit naturally in one document. |
| 2 | Hook event deduplication | **Works.** A unique index on `(workflow_id, external_id)` rejects the duplicate with `DuplicateKeyError`. |
| 3 | Dangling reference to a user | **Not protected.** A workflow pointing to a non-existent user was accepted: MongoDB has no foreign keys, so integrity would live in our code. |
| 4 | Multi-document transaction | **Refused** on the default container: `Transaction numbers are only allowed on a replica set member or mongos`. Transactions need a replica set, even with one node. |

## What consumed the most minutes

Nothing blocked. The only real cost is the image: `mongo:8.0` is **1.29 GB** on disk.

## Conclusions for the persistence decision

- MongoDB is good at the **workflow definitions** (nested, variable length) and at **deduplication**.
- It is weak at what is **relational** in AREA: users, linked identities, OAuth tokens and service subscriptions all reference each other, and check 3 shows that nothing stops orphans.
- Atomic "run + event processed" writes (needed by ENG-RETRY idempotency) require a **replica set** in Docker Compose: more configuration and one more thing to health-check.
- A relational database with a JSON column (e.g. PostgreSQL `jsonb`) could cover both sides. This should be checked against the PostgreSQL spike rather than assumed.

## What I would need to go further

- Configure a single-node replica set in Compose and re-run check 4.
- Enable authentication: the default `mongo` image starts **without any authentication** unless root credentials are provided.
- Test schema validation (`$jsonSchema`) as a partial replacement for missing constraints.

## How to reproduce

```bash
cd spikes/mongodb
docker compose up -d
python -m venv .venv
.venv/Scripts/python -m pip install -r requirements.txt   # Linux/macOS: .venv/bin/python
MONGO_URL="mongodb://localhost:27017" .venv/Scripts/python spike_mongo.py
docker compose down
```

Expected output:

```
[1] workflow stored and read back in one document: 4 reactions, condition kept
[2] duplicate hook event rejected by the unique index
[3] workflow pointing to a NON-EXISTENT user was accepted: no foreign keys in MongoDB
[4] transaction REFUSED on a standalone container: Transaction numbers are only allowed on a replica set member or mongos
```
