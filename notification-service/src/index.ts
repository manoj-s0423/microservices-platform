import { createApp } from './app';
import { config } from './config';
import { logger } from './logger';
import { connectDatabase, disconnectDatabase } from './db';

async function main() {
  try {
    await connectDatabase();
    logger.info('connected to mongodb');
  } catch (err) {
    // Database connection failure / missing MONGO_URI - fail fast rather
    // than serving traffic that can never actually persist anything.
    logger.fatal({ err: (err as Error).message }, 'failed to connect to mongodb, exiting');
    process.exit(1);
  }

  const app = createApp();
  const server = app.listen(config.port, () => {
    logger.info({ port: config.port, env: config.env }, 'notification-service listening');
  });

  async function shutdown(signal: string) {
    logger.info({ signal }, 'shutdown signal received, closing server');

    const forceExit = setTimeout(() => {
      logger.error('forced shutdown after timeout');
      process.exit(1);
    }, config.shutdownTimeoutMs);
    forceExit.unref();

    server.close(async () => {
      await disconnectDatabase();
      logger.info('server closed, exiting');
      process.exit(0);
    });
  }

  process.on('SIGTERM', () => shutdown('SIGTERM'));
  process.on('SIGINT', () => shutdown('SIGINT'));
}

process.on('unhandledRejection', (reason) => {
  logger.error({ reason }, 'unhandled promise rejection');
});

process.on('uncaughtException', (err) => {
  logger.fatal({ err: err.message, stack: err.stack }, 'uncaught exception, exiting');
  process.exit(1);
});

main();
