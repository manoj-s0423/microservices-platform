"""SQLAlchemy engine/session setup.

The engine is created with `pool_pre_ping=True` so a stale/dropped
connection (e.g. after the DB restarts) is detected and recycled instead of
surfacing as a confusing mid-request error - one of the "database
connection failure" scenarios this service is meant to help you practice
diagnosing (the other being a wrong DB_HOST/DB_PASSWORD at startup, which
fails immediately on first request/health check instead).
"""
from collections.abc import Generator

from sqlalchemy import create_engine
from sqlalchemy.orm import DeclarativeBase, Session, sessionmaker

from app.config import get_settings

settings = get_settings()

engine = create_engine(
    settings.database_url,
    pool_size=settings.db_pool_size,
    max_overflow=settings.db_pool_max_overflow,
    pool_pre_ping=True,
    connect_args={"connect_timeout": settings.db_connect_timeout_seconds},
)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


class Base(DeclarativeBase):
    pass


def get_db() -> Generator[Session, None, None]:
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
