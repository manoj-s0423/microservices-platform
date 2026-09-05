'use strict';

require('dotenv').config();

function requireEnv(name, fallback) {
  const value = process.env[name] ?? fallback;
  if (value === undefined) {
    // Intentionally do not throw at import time for every optional var;
    // JWT_SECRET is validated separately at startup (see index.js) so that
    // "incorrect/missing environment variable" is a reproducible, isolated
    // failure scenario rather than a crash-on-import surprise.
    return undefined;
  }
  return value;
}

const config = {
  env: process.env.NODE_ENV || 'development',
  port: parseInt(process.env.PORT || '3000', 10),
  logLevel: process.env.LOG_LEVEL || 'info',

  jwtSecret: requireEnv('JWT_SECRET'),
  // The `cors` package treats an array as an exact-match allowlist - an
  // array containing the string '*' means "allow an Origin header that
  // is literally the four characters *", which no real browser ever
  // sends, NOT "allow everything". Passing the bare string '*' (not
  // wrapped in an array) is what actually triggers cors' wildcard
  // behavior. So the default case and the explicit-whitelist case need
  // different shapes, not just different values.
  corsAllowedOrigins:
    !process.env.CORS_ALLOWED_ORIGINS || process.env.CORS_ALLOWED_ORIGINS === '*'
      ? '*'
      : process.env.CORS_ALLOWED_ORIGINS.split(',').map((s) => s.trim()),

  rateLimit: {
    windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '60000', 10),
    max: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100', 10),
  },

  services: {
    user: process.env.USER_SERVICE_URL || 'http://localhost:8081',
    product: process.env.PRODUCT_SERVICE_URL || 'http://localhost:8000',
    order: process.env.ORDER_SERVICE_URL || 'http://localhost:8082',
    payment: process.env.PAYMENT_SERVICE_URL || 'http://localhost:8083',
    notification: process.env.NOTIFICATION_SERVICE_URL || 'http://localhost:8084',
  },

  // Every service's liveness endpoint, keyed the same as `services` above.
  // user-service is Spring Boot and uses Actuator's `/actuator/health`
  // convention rather than the plain `/health` the other four services
  // expose - this map is what lets /ready poll each service correctly
  // instead of assuming one path fits all languages/frameworks.
  serviceHealthPaths: {
    user: '/actuator/health',
    product: '/health',
    order: '/health',
    payment: '/health',
    notification: '/health',
  },

  http: {
    timeoutMs: parseInt(process.env.HTTP_TIMEOUT_MS || '3000', 10),
    retryAttempts: parseInt(process.env.HTTP_RETRY_ATTEMPTS || '2', 10),
    retryDelayMs: parseInt(process.env.HTTP_RETRY_DELAY_MS || '200', 10),
  },

  shutdownTimeoutMs: parseInt(process.env.SHUTDOWN_TIMEOUT_MS || '10000', 10),
};

module.exports = config;
