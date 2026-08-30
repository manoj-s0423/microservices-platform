import uuid

from sqlalchemy import func, select
from sqlalchemy.exc import IntegrityError
from sqlalchemy.orm import Session

from app.models import Product
from app.schemas.product import ProductCreate, ProductUpdate


class ProductNotFoundError(Exception):
    def __init__(self, product_id: uuid.UUID):
        self.product_id = product_id
        super().__init__(f"Product {product_id} not found")


class DuplicateSkuError(Exception):
    def __init__(self, sku: str):
        self.sku = sku
        super().__init__(f"Product with SKU '{sku}' already exists")


def list_products(
    db: Session,
    *,
    category: str | None = None,
    active_only: bool = True,
    page: int = 1,
    page_size: int = 20,
) -> tuple[list[Product], int]:
    query = select(Product)
    if category:
        query = query.where(Product.category == category)
    if active_only:
        query = query.where(Product.is_active.is_(True))

    total = db.scalar(select(func.count()).select_from(query.subquery())) or 0

    items = (
        db.execute(
            query.order_by(Product.created_at.desc()).offset((page - 1) * page_size).limit(page_size)
        )
        .scalars()
        .all()
    )
    return list(items), total


def get_product(db: Session, product_id: uuid.UUID) -> Product:
    product = db.get(Product, product_id)
    if product is None:
        raise ProductNotFoundError(product_id)
    return product


def create_product(db: Session, data: ProductCreate) -> Product:
    product = Product(**data.model_dump())
    db.add(product)
    try:
        db.commit()
    except IntegrityError as exc:
        db.rollback()
        raise DuplicateSkuError(data.sku) from exc
    db.refresh(product)
    return product


def update_product(db: Session, product_id: uuid.UUID, data: ProductUpdate) -> Product:
    product = get_product(db, product_id)
    for field, value in data.model_dump(exclude_unset=True).items():
        setattr(product, field, value)
    db.commit()
    db.refresh(product)
    return product


def delete_product(db: Session, product_id: uuid.UUID) -> None:
    product = get_product(db, product_id)
    product.is_active = False  # soft delete: preserves referential history for orders
    db.commit()
