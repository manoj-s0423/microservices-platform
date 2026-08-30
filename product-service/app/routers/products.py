import uuid

from fastapi import APIRouter, Depends, Query, status
from sqlalchemy.orm import Session

from app.config import get_settings
from app.database import get_db
from app.schemas.product import (
    ProductCreate,
    ProductListResponse,
    ProductResponse,
    ProductUpdate,
)
from app.services import product_service

router = APIRouter(prefix="/api/v1/products", tags=["products"])
settings = get_settings()

# NOTE: ProductNotFoundError / DuplicateSkuError are intentionally left
# uncaught here - they propagate to the exception handlers registered in
# main.py, which is the single place response shape ({"error", "message"})
# is decided. Catching-and-rethrowing as HTTPException here would bypass
# those handlers and produce an inconsistent {"detail": ...} body instead.


@router.get("", response_model=ProductListResponse)
def list_products(
    category: str | None = None,
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=settings.default_page_size, ge=1, le=settings.max_page_size),
    db: Session = Depends(get_db),
):
    items, total = product_service.list_products(db, category=category, page=page, page_size=page_size)
    return ProductListResponse(
        items=[ProductResponse.model_validate(p) for p in items],
        total=total,
        page=page,
        page_size=page_size,
    )


@router.get("/{product_id}", response_model=ProductResponse)
def get_product(product_id: uuid.UUID, db: Session = Depends(get_db)):
    return product_service.get_product(db, product_id)


@router.post("", response_model=ProductResponse, status_code=status.HTTP_201_CREATED)
def create_product(payload: ProductCreate, db: Session = Depends(get_db)):
    return product_service.create_product(db, payload)


@router.patch("/{product_id}", response_model=ProductResponse)
def update_product(product_id: uuid.UUID, payload: ProductUpdate, db: Session = Depends(get_db)):
    return product_service.update_product(db, product_id, payload)


@router.delete("/{product_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_product(product_id: uuid.UUID, db: Session = Depends(get_db)):
    product_service.delete_product(db, product_id)
