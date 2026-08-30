import dotenv from 'dotenv';

dotenv.config();

export interface Config {
  env: string;
  port: number;
  logLevel: string;
  shutdownTimeoutMs: number;

  mongoUri: string;
  mongoConnectTimeoutMs: number;

  emailProvider: {
    mode: 'simulated' | 'live';
    apiUrl: string;
    apiKey: string;
    timeoutMs: number;
    retryAttempts: number;
  };

  smsProvider: {
    mode: 'simulated' | 'live';
    apiUrl: string;
    apiKey: string;
  };
}

function envInt(name: string, fallback: number): number {
  const raw = process.env[name];
  if (!raw) return fallback;
  const parsed = parseInt(raw, 10);
  return Number.isNaN(parsed) ? fallback : parsed;
}

export const config: Config = {
  env: process.env.NODE_ENV || 'development',
  port: envInt('PORT', 8084),
  logLevel: process.env.LOG_LEVEL || 'info',
  shutdownTimeoutMs: envInt('SHUTDOWN_TIMEOUT_MS', 10000),

  // Deliberately no fallback for MONGO_URI: an empty connection string
  // fails fast and loudly at startup rather than silently trying
  // "mongodb://localhost:27017" in a production-like environment.
  mongoUri: process.env.MONGO_URI || '',
  mongoConnectTimeoutMs: envInt('MONGO_CONNECT_TIMEOUT_MS', 5000),

  emailProvider: {
    mode: (process.env.EMAIL_PROVIDER_MODE as 'simulated' | 'live') || 'simulated',
    apiUrl: process.env.EMAIL_PROVIDER_API_URL || '',
    apiKey: process.env.EMAIL_PROVIDER_API_KEY || '',
    timeoutMs: envInt('EMAIL_PROVIDER_TIMEOUT_MS', 3000),
    retryAttempts: envInt('EMAIL_PROVIDER_RETRY_ATTEMPTS', 2),
  },

  smsProvider: {
    mode: (process.env.SMS_PROVIDER_MODE as 'simulated' | 'live') || 'simulated',
    apiUrl: process.env.SMS_PROVIDER_API_URL || '',
    apiKey: process.env.SMS_PROVIDER_API_KEY || '',
  },
};
