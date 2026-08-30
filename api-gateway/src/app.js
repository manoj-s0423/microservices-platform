'use strict';

const express = require('express');
const helmet = require('helmet');
const cors = require('cors');
const rateLimit = require('express-rate-limit');

const config = require('./config');
const requestLogger = require('./middleware/requestLogger');
const { errorHandler, notFoundHandler } = require('./middleware/errorHandler');

const healthRoutes = require('./routes/health');
const userRoutes = require('./routes/users');
const productRoutes = require('./routes/products');
const orderRoutes = require('./routes/orders');

function createApp() {
  const app = express();

  app.disable('x-powered-by');
  app.use(helmet());
  app.use(cors({ origin: config.corsAllowedOrigins }));
  app.use(express.json({ limit: '1mb' }));
  app.use(requestLogger);

  app.use(
    rateLimit({
      windowMs: config.rateLimit.windowMs,
      max: config.rateLimit.max,
      standardHeaders: true,
      legacyHeaders: false,
      message: { error: 'rate_limited', message: 'Too many requests, please try again later.' },
    })
  );

  // Health/readiness are unversioned and unauthenticated by convention.
  app.use('/', healthRoutes);

  app.use('/api/v1', userRoutes);
  app.use('/api/v1', productRoutes);
  app.use('/api/v1', orderRoutes);

  app.use(notFoundHandler);
  app.use(errorHandler);

  return app;
}

module.exports = createApp;
