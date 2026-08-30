import time
import uuid

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
import structlog

from app.config import get_settings
from app.logging_config import configure_logging, get_logger
from app.routers import health, products
from app.services.product_service import DuplicateSkuError, ProductNotFoundError

settings = get_settings()
configure_logging(settings.log_level)
logger = get_logger()

app = FastAPI(
    title="ShopStream Product Service",
    description="Owns the product catalog: CRUD, search, and stock levels.",
    version="1.0.0",
)


@app.middleware("http")
async def request_context_middleware(request: Request, call_next):
    """Propagates X-Request-Id for tracing and logs latency per request."""
    request_id = request.headers.get("x-request-id", str(uuid.uuid4()))
    structlog.contextvars.bind_contextvars(request_id=request_id)
    start = time.perf_counter()
    response = await call_next(request)
    duration_ms = (time.perf_counter() - start) * 1000
    response.headers["x-request-id"] = request_id
    logger.info(
        "request_completed",
        method=request.method,
        path=request.url.path,
        status_code=response.status_code,
        duration_ms=round(duration_ms, 2),
    )
    structlog.contextvars.clear_contextvars()
    return response


@app.exception_handler(ProductNotFoundError)
async def not_found_handler(request: Request, exc: ProductNotFoundError):
    return JSONResponse(status_code=404, content={"error": "product_not_found", "message": str(exc)})


@app.exception_handler(DuplicateSkuError)
async def duplicate_sku_handler(request: Request, exc: DuplicateSkuError):
    return JSONResponse(status_code=409, content={"error": "duplicate_sku", "message": str(exc)})


app.include_router(health.router)
app.include_router(products.router)


@app.get("/")
def root():
    return {"service": "product-service", "status": "running"}
