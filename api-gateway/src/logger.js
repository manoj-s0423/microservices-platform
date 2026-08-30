'use strict';

const pino = require('pino');
const config = require('./config');

// Structured (JSON) logging so log aggregators (e.g. Loki/CloudWatch/ELK)
// can parse fields directly. In development, pretty-print for readability
// if pino-pretty is installed; otherwise fall back to raw JSON.
const logger = pino({
  level: config.logLevel,
  base: { service: 'api-gateway' },
  timestamp: pino.stdTimeFunctions.isoTime,
});

module.exports = logger;
