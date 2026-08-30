from datetime import datetime, timezone

from fastapi import APIRouter, Depends, Response, status
from sqlalchemy import text
from sqlalchemy.exc import OperationalError
from sqlalchemy.orm import Session

from app.database import get_db

router = APIRouter(tags=["health"])


@router.get("/health")
def health():
    """Liveness: process is up. No dependency checks on purpose."""
    return {"status": "UP", "service": "product-service", "timestamp": datetime.now(timezone.utc).isoformat()}


@router.get("/ready")
def ready(response: Response, db: Session = Depends(get_db)):
    """Readiness: can we actually serve traffic (i.e. reach the database)?"""
    try:
        db.execute(text("SELECT 1"))
        return {"status": "UP", "dependencies": {"database": "UP"}}
    except OperationalError:
        response.status_code = status.HTTP_503_SERVICE_UNAVAILABLE
        return {"status": "DEGRADED", "dependencies": {"database": "DOWN"}}
