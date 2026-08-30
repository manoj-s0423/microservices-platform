'use strict';

const { randomUUID } = require('crypto');
const pinoHttp = require('pino-http');
const logger = require('../logger');

/**
 * Attaches/propagates an X-Request-Id for distributed tracing across
 * services, and logs each request/response with latency in structured form.
 */
const requestLogger = pinoHttp({
  logger,
  genReqId: (req, res) => {
    const existing = req.headers['x-request-id'];
    const id = existing || randomUUID();
    res.setHeader('x-request-id', id);
    return id;
  },
  customLogLevel: (req, res, err) => {
    if (err || res.statusCode >= 500) return 'error';
    if (res.statusCode >= 400) return 'warn';
    return 'info';
  },
});

module.exports = requestLogger;
