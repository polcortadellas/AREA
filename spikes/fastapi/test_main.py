from fastapi.testclient import TestClient

from main import app

client = TestClient(app)


def test_ping_returns_exact_body():
    response = client.get("/ping")
    assert response.status_code == 200
    assert response.text == '{"status":"ok"}'


def test_ping_rejects_other_methods():
    response = client.post("/ping")
    assert response.status_code == 405
