'use strict';

const express = require('express');
const config = require('../config');
const { createServiceClient } = require('../services/httpClient');
const { authenticate } = require('../middleware/auth');

const router = express.Router();
const userClient = createServiceClient(config.services.user, 'user-service');

// Public: login is proxied without requiring an existing token.
router.post('/auth/login', async (req, res, next) => {
  try {
    const { data, status } = await userClient.post('/api/v1/auth/login', req.body);
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.post('/auth/register', async (req, res, next) => {
  try {
    const { data, status } = await userClient.post('/api/v1/auth/register', req.body);
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.get('/users/me', authenticate, async (req, res, next) => {
  try {
    const { data, status } = await userClient.get(`/api/v1/users/${req.user.sub}`, {
      headers: { 'x-request-id': req.id },
    });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.get('/users/:id', authenticate, async (req, res, next) => {
  try {
    const { data, status } = await userClient.get(`/api/v1/users/${req.params.id}`, {
      headers: { 'x-request-id': req.id },
    });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

module.exports = router;
