'use strict';

const express = require('express');
const config = require('../config');
const { createServiceClient } = require('../services/httpClient');
const { authenticate } = require('../middleware/auth');

const router = express.Router();
const orderClient = createServiceClient(config.services.order, 'order-service');

router.post('/orders', authenticate, async (req, res, next) => {
  try {
    const payload = { ...req.body, userId: req.user.sub };
    const { data, status } = await orderClient.post('/api/v1/orders', payload, {
      headers: { 'x-request-id': req.id },
    });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.get('/orders/:id', authenticate, async (req, res, next) => {
  try {
    const { data, status } = await orderClient.get(`/api/v1/orders/${req.params.id}`, {
      headers: { 'x-request-id': req.id },
    });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.get('/orders', authenticate, async (req, res, next) => {
  try {
    const { data, status } = await orderClient.get('/api/v1/orders', {
      params: { userId: req.user.sub },
      headers: { 'x-request-id': req.id },
    });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

module.exports = router;
