# Recommendation

**Chosen stack:** Go (standard library) for the server, PostgreSQL for persistence, React for the web client and React Native for the mobile client.

**The two criteria that drove the decision.** First, concurrency for hooks: AREA mostly polls many external services in parallel, and Go's goroutines fit that better than NestJS's single-threaded scheduler, which needed a manual guard against overlapping runs. The Go image is also 21 MB, against 401 MB for NestJS. Second, the data model: users, identities and workflows reference each other, while each service has its own flexible configuration. PostgreSQL gives foreign keys, `jsonb`, real transactions and unique constraints for idempotency. MongoDB accepted a workflow owned by a user that does not exist, and SQLite allows a single writer.

**Trade-off we accept.** Nobody in the team knew Go (learning curve scored 2), and the spikes did not test OpenAPI generation or OAuth2 from Go. We will test both first in the next milestone and keep FastAPI as the fallback. The web matrix ranked Vue (60) above React (52), but the subject asks for a mobile client, and with React we can use React Native: the same language and the same API client generated from the OpenAPI description. The matrix has no criterion for that, so we will add one.

Runs are queued through a transactional outbox in PostgreSQL and Redis Streams as the broker ([ADR 0003](../docs/adr/0003-workflow-engine-broker.md)); Redis Streams is not yet validated in a spike, which task M-11 will do.

**Disagreement inside the team and how it was settled.** [TO FILL: describe the real disagreement and how it was resolved.]
