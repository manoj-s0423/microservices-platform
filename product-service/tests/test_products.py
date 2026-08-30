def test_create_product_returns_201(client, sample_product_payload):
    res = client.post("/api/v1/products", json=sample_product_payload)
    assert res.status_code == 201
    body = res.json()
    assert body["sku"] == sample_product_payload["sku"]
    assert body["is_active"] is True


def test_create_product_rejects_duplicate_sku(client, sample_product_payload):
    client.post("/api/v1/products", json=sample_product_payload)
    res = client.post("/api/v1/products", json=sample_product_payload)
    assert res.status_code == 409
    assert res.json()["error"] == "duplicate_sku"


def test_create_product_rejects_invalid_price(client, sample_product_payload):
    sample_product_payload["price_cents"] = -100
    res = client.post("/api/v1/products", json=sample_product_payload)
    assert res.status_code == 422


def test_get_product_by_id(client, sample_product_payload):
    created = client.post("/api/v1/products", json=sample_product_payload).json()
    res = client.get(f"/api/v1/products/{created['id']}")
    assert res.status_code == 200
    assert res.json()["name"] == "Test Widget"


def test_get_missing_product_returns_404(client):
    res = client.get("/api/v1/products/00000000-0000-0000-0000-000000000000")
    assert res.status_code == 404
    assert res.json()["error"] == "product_not_found"


def test_list_products_paginates(client, sample_product_payload):
    for i in range(3):
        payload = dict(sample_product_payload)
        payload["sku"] = f"SKU-TEST-{i:04d}"
        client.post("/api/v1/products", json=payload)

    res = client.get("/api/v1/products", params={"page": 1, "page_size": 2})
    body = res.json()
    assert res.status_code == 200
    assert body["total"] == 3
    assert len(body["items"]) == 2


def test_list_products_filters_by_category(client, sample_product_payload):
    client.post("/api/v1/products", json=sample_product_payload)
    other = dict(sample_product_payload, sku="SKU-OTHER-0001", category="other-category")
    client.post("/api/v1/products", json=other)

    res = client.get("/api/v1/products", params={"category": "test"})
    body = res.json()
    assert all(item["category"] == "test" for item in body["items"])


def test_update_product_partial_fields(client, sample_product_payload):
    created = client.post("/api/v1/products", json=sample_product_payload).json()
    res = client.patch(f"/api/v1/products/{created['id']}", json={"stock_quantity": 42})
    assert res.status_code == 200
    assert res.json()["stock_quantity"] == 42
    assert res.json()["name"] == sample_product_payload["name"]  # unchanged


def test_delete_product_soft_deletes(client, sample_product_payload):
    created = client.post("/api/v1/products", json=sample_product_payload).json()
    res = client.delete(f"/api/v1/products/{created['id']}")
    assert res.status_code == 204

    # Soft-deleted products are excluded from the default active-only listing.
    listing = client.get("/api/v1/products").json()
    assert all(item["id"] != created["id"] for item in listing["items"])
