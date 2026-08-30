import mongoose from 'mongoose';
import { config } from './config';
import { logger } from './logger';

export async function connectDatabase(): Promise<typeof mongoose> {
  if (!config.mongoUri) {
    // Incorrect/missing environment variable - fail fast and loud instead
    // of hanging in mongoose's default connection retry loop.
    throw new Error('MONGO_URI is not configured; refusing to start');
  }

  mongoose.connection.on('disconnected', () => {
    logger.warn('mongodb connection lost');
  });
  mongoose.connection.on('reconnected', () => {
    logger.info('mongodb connection restored');
  });

  return mongoose.connect(config.mongoUri, {
    serverSelectionTimeoutMS: config.mongoConnectTimeoutMs,
  });
}

export function isDatabaseConnected(): boolean {
  // 1 === connected, per mongoose.ConnectionStates
  return mongoose.connection.readyState === 1;
}

export async function disconnectDatabase(): Promise<void> {
  await mongoose.disconnect();
}
