import os

import uvicorn
from fastapi import FastAPI

app = FastAPI(title="AREA spike - FastAPI")


@app.get("/ping")
def ping() -> dict[str, str]:
    return {"status": "ok"}


if __name__ == "__main__":
    # No default value on purpose: the port must come from the environment.
    raw_port = os.environ.get("PORT")
    if raw_port is None:
        raise SystemExit("PORT environment variable is required")
    uvicorn.run(app, host="0.0.0.0", port=int(raw_port))
