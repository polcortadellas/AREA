"""MongoDB spike: four checks that matter for AREA's persistence layer.

1. Store a whole workflow (trigger + condition + reactions) as one document.
2. Reject the same hook event twice (deduplication through a unique index).
3. See whether a workflow can reference a user that does not exist.
4. See whether multi-document transactions work on a default container.
"""

import os
import sys

from bson import ObjectId
from pymongo import MongoClient
from pymongo.errors import DuplicateKeyError, OperationFailure

mongo_url = os.environ.get("MONGO_URL")
if mongo_url is None:
    sys.exit("MONGO_URL environment variable is required")

client = MongoClient(mongo_url, serverSelectionTimeoutMS=5000)
db = client["area_spike"]
db.users.drop()
db.workflows.drop()
db.hook_events.drop()

# 1. A workflow is a nested structure: one document holds the whole tree.
user_id = db.users.insert_one({"email": "demo@example.com"}).inserted_id
workflow = {
    "owner_id": user_id,
    "name": "Invoice received",
    "enabled": True,
    "trigger": {"service": "gmail", "action": "new_mail_with_attachment"},
    "condition": {"left": "{{ trigger.amount }}", "op": "<", "right": 10000},
    "then": [
        {"service": "drive", "reaction": "save_file", "params": {"file": "{{ trigger.file }}"}},
        {"service": "accounting", "reaction": "book_invoice", "params": {}},
    ],
    "else": [
        {"service": "teams", "reaction": "ask_validation", "params": {}},
        {"service": "slack", "reaction": "alert_cfo", "params": {}},
    ],
}
workflow_id = db.workflows.insert_one(workflow).inserted_id
stored = db.workflows.find_one({"_id": workflow_id})
reactions = len(stored["then"]) + len(stored["else"])
print(f"[1] workflow stored and read back in one document: {reactions} reactions, condition kept")

# 2. Hook deduplication: the pair (workflow, external event id) must be unique.
db.hook_events.create_index([("workflow_id", 1), ("external_id", 1)], unique=True)
event = {"workflow_id": workflow_id, "external_id": "gmail-msg-INV-2291"}
db.hook_events.insert_one(dict(event))
try:
    db.hook_events.insert_one(dict(event))
    print("[2] FAIL: duplicate event accepted")
except DuplicateKeyError:
    print("[2] duplicate hook event rejected by the unique index")

# 3. Referential integrity: is a dangling owner_id refused?
ghost_user = ObjectId()
db.workflows.insert_one({"owner_id": ghost_user, "name": "orphan"})
print("[3] workflow pointing to a NON-EXISTENT user was accepted: no foreign keys in MongoDB")

# 4. Multi-document transaction (e.g. create run record + mark event processed).
try:
    with client.start_session() as session:
        with session.start_transaction():
            db.hook_events.insert_one({"workflow_id": workflow_id, "external_id": "tx-test"}, session=session)
            db.workflows.update_one({"_id": workflow_id}, {"$set": {"last_run": "now"}}, session=session)
    print("[4] transaction committed")
except OperationFailure as error:
    print(f"[4] transaction REFUSED on a standalone container: {error.details.get('errmsg')}")
