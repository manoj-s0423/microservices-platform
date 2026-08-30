import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from app.database import Base, get_db
from app.main import app

# Tests run against an in-memory SQLite DB so `pytest` needs no external
# PostgreSQL instance. Business logic (services/product_service.py) is
# database-agnostic SQLAlchemy Core/ORM, so this is a faithful substitute
# for unit/API-level tests; real cross-service integration still targets
# PostgreSQL (see README "Test" section).
engine = create_engine(
    "sqlite:///:memory:",
    connect_args={"check_same_thread": False},
    poolclass=StaticPool,
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


@pytest.fixture()
def db_session():
    Base.metadata.create_all(bind=engine)
    session = TestingSessionLocal()
    try:
        yield session
    finally:
        session.close()
        Base.metadata.drop_all(bind=engine)


@pytest.fixture()
def client(db_session):
    def override_get_db():
        yield db_session

    app.dependency_overrides[get_db] = override_get_db
    with TestClient(app) as test_client:
        yield test_client
    app.dependency_overrides.clear()


@pytest.fixture()
def sample_product_payload():
    return {
        "sku": "SKU-TEST-0001",
        "name": "Test Widget",
        "description": "A widget used only in tests.",
        "category": "test",
        "price_cents": 1500,
        "currency": "USD",
        "stock_quantity": 10,
    }
