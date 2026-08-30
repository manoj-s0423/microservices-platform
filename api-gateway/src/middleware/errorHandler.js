'use strict';

const logger = require('../logger');

/**
 * Central error handler. Normalizes axios errors from downstream calls
 * into consistent gateway responses so failure scenarios are observable
 * and distinguishable at the edge:
 *  - ECONNREFUSED / ENOTFOUND -> 502 (service unavailable / DNS failure)
 *  - ECONNABORTED (timeout)   -> 504 (slow downstream / timeout)
 *  - downstream 4xx/5xx       -> proxied through with original status
 */
function errorHandler(err, req, res, _next) {
  const requestId = req.id;

  if (err.isAxiosError) {
    if (err.response) {
      logger.warn({ requestId, status: err.response.status, url: err.config?.url }, 'downstream returned error');
      return res.status(err.response.status).json(err.response.data);
    }
    if (err.code === 'ECONNABORTED') {
      logger.error({ requestId, url: err.config?.url }, 'downstream timeout');
      return res.status(504).json({ error: 'gateway_timeout', message: 'Downstream service did not respond in time' });
    }
    logger.error({ requestId, url: err.config?.url, code: err.code }, 'downstream unreachable');
    return res.status(502).json({ error: 'bad_gateway', message: 'Downstream service is unreachable' });
  }

  logger.error({ requestId, err: err.message, stack: err.stack }, 'unhandled error');
  return res.status(err.status || 500).json({
    error: 'internal_error',
    message: config_isProd() ? 'An unexpected error occurred' : err.message,
  });
}

function config_isProd() {
  return process.env.NODE_ENV === 'production';
}

function notFoundHandler(req, res) {
  res.status(404).json({ error: 'not_found', message: `Route ${req.method} ${req.path} not found` });
}

module.exports = { errorHandler, notFoundHandler };
