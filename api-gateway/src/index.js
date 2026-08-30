'use strict';

const createApp = require('./app');
const config = require('./config');
const logger = require('./logger');

const app = createApp();

const server = app.listen(config.port, () => {
  logger.info({ port: config.port, env: config.env }, 'api-gateway listening');
});

// Graceful shutdown: stop accepting new connections, let in-flight requests
// finish, then exit. This is what lets you practice "pod killed mid-request"
// vs. "clean rolling deploy" as two distinct, observable behaviors.
function shutdown(signal) {
  logger.info({ signal }, 'shutdown signal received, closing server');
  server.close(() => {
    logger.info('server closed, exiting');
    process.exit(0);
  });

  setTimeout(() => {
    logger.error('forced shutdown after timeout');
    process.exit(1);
  }, config.shutdownTimeoutMs).unref();
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));

process.on('unhandledRejection', (reason) => {
  logger.error({ reason }, 'unhandled promise rejection');
});

process.on('uncaughtException', (err) => {
  logger.fatal({ err: err.message, stack: err.stack }, 'uncaught exception, exiting');
  process.exit(1);
});

module.exports = server;
