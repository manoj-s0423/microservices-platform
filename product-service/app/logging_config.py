"""Structured (JSON) logging configuration using structlog.

Every log line carries service name, level, timestamp, and a bound
request_id (see the request-id middleware in main.py) so logs can be
correlated across the API Gateway -> product-service hop and ingested by
a log aggregator without regex parsing.
"""
import logging
import sys

import structlog


def configure_logging(log_level: str = "info") -> None:
    logging.basicConfig(
        format="%(message)s",
        stream=sys.stdout,
        level=getattr(logging, log_level.upper(), logging.INFO),
    )

    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.JSONRenderer(),
        ],
        logger_factory=structlog.stdlib.LoggerFactory(),
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(logging, log_level.upper(), logging.INFO)
        ),
        cache_logger_on_first_use=True,
    )


def get_logger(name: str = "product-service"):
    return structlog.get_logger(name)
