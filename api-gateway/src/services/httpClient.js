'use strict';

const axios = require('axios');
const config = require('../config');
const logger = require('../logger');

/**
 * Creates an axios instance pre-configured with a timeout and bounded
 * exponential-backoff retries for transient failures (network errors,
 * timeouts, and 502/503/504). This is the mechanism that lets you
 * reproduce/practice:
 *   - slow API response (downstream sleeps -> gateway times out)
 *   - failed service-to-service communication (downstream down -> retries then fails)
 *   - DNS/service discovery problems (unresolved host -> ENOTFOUND, no pointless retry storm)
 */
function createServiceClient(baseURL, serviceName) {
  const client = axios.create({
    baseURL,
    timeout: config.http.timeoutMs,
    validateStatus: (status) => status < 500, // let 5xx flow to error handling/retry
  });

  const isRetryable = (error) => {
    if (!error.response) return true; // network error, timeout, DNS failure
    return [502, 503, 504].includes(error.response.status);
  };

  client.interceptors.response.use(
    (response) => response,
    async (error) => {
      const cfg = error.config || {};
      cfg.__retryCount = cfg.__retryCount || 0;

      if (cfg.__retryCount < config.http.retryAttempts && isRetryable(error)) {
        cfg.__retryCount += 1;
        const delay = config.http.retryDelayMs * 2 ** (cfg.__retryCount - 1);
        logger.warn(
          {
            service: serviceName,
            attempt: cfg.__retryCount,
            delayMs: delay,
            url: cfg.url,
            error: error.message,
          },
          'retrying downstream call'
        );
        await new Promise((resolve) => setTimeout(resolve, delay));
        return client(cfg);
      }

      logger.error(
        { service: serviceName, url: cfg.url, error: error.message },
        'downstream call failed'
      );
      return Promise.reject(error);
    }
  );

  return client;
}

module.exports = { createServiceClient };
