def test_health_returns_up(client):
    res = client.get("/health")
    assert res.status_code == 200
    body = res.json()
    assert body["status"] == "UP"
    assert body["service"] == "product-service"


def test_ready_returns_up_when_db_reachable(client):
    res = client.get("/ready")
    assert res.status_code == 200
    assert res.json()["dependencies"]["database"] == "UP"
