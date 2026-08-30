'use strict';

const jwt = require('jsonwebtoken');
const config = require('../config');

/**
 * Verifies a bearer JWT issued by user-service. The gateway does not issue
 * tokens itself; it validates them on behalf of downstream services so
 * routes that don't need identity (health checks, public catalog reads)
 * can skip auth entirely.
 */
function authenticate(req, res, next) {
  const header = req.headers.authorization || '';
  const [scheme, token] = header.split(' ');

  if (scheme !== 'Bearer' || !token) {
    return res.status(401).json({ error: 'missing_token', message: 'Authorization header must be: Bearer <token>' });
  }

  if (!config.jwtSecret) {
    // Reproducible "incorrect environment variable" failure scenario:
    // if JWT_SECRET was never set, every authenticated route fails clearly
    // instead of silently accepting/rejecting tokens.
    return res.status(500).json({ error: 'server_misconfigured', message: 'JWT_SECRET is not configured' });
  }

  try {
    req.user = jwt.verify(token, config.jwtSecret);
    return next();
  } catch (err) {
    return res.status(401).json({ error: 'invalid_token', message: err.message });
  }
}

module.exports = { authenticate };
