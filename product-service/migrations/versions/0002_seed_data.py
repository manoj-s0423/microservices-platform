"""seed sample catalog data for local development

Revision ID: 0002
Revises: 0001
Create Date: 2026-01-01 00:05:00
"""
import uuid

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

revision = "0002"
down_revision = "0001"
branch_labels = None
depends_on = None

products_table = sa.table(
    "products",
    sa.column("id", postgresql.UUID(as_uuid=True)),
    sa.column("sku", sa.String),
    sa.column("name", sa.String),
    sa.column("description", sa.String),
    sa.column("category", sa.String),
    sa.column("price_cents", sa.Integer),
    sa.column("currency", sa.String),
    sa.column("stock_quantity", sa.Integer),
)

SEED_PRODUCTS = [
    {
        "id": uuid.UUID("aaaaaaaa-0000-0000-0000-000000000001"),
        "sku": "SKU-WIRELESS-MOUSE",
        "name": "Wireless Mouse",
        "description": "Ergonomic 2.4GHz wireless mouse with USB receiver.",
        "category": "electronics",
        "price_cents": 1999,
        "currency": "USD",
        "stock_quantity": 250,
    },
    {
        "id": uuid.UUID("aaaaaaaa-0000-0000-0000-000000000002"),
        "sku": "SKU-MECH-KEYBOARD",
        "name": "Mechanical Keyboard",
        "description": "Tactile mechanical keyboard, hot-swappable switches.",
        "category": "electronics",
        "price_cents": 8999,
        "currency": "USD",
        "stock_quantity": 80,
    },
    {
        "id": uuid.UUID("aaaaaaaa-0000-0000-0000-000000000003"),
        "sku": "SKU-STEEL-BOTTLE",
        "name": "Insulated Steel Bottle",
        "description": "24oz double-walled insulated bottle.",
        "category": "home",
        "price_cents": 2499,
        "currency": "USD",
        "stock_quantity": 500,
    },
]


def upgrade() -> None:
    op.bulk_insert(products_table, SEED_PRODUCTS)


def downgrade() -> None:
    conn = op.get_bind()
    ids = [str(p["id"]) for p in SEED_PRODUCTS]
    conn.execute(sa.text("DELETE FROM products WHERE id = ANY(:ids)"), {"ids": ids})
