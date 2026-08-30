'use strict';

const express = require('express');
const config = require('../config');
const { createServiceClient } = require('../services/httpClient');

const router = express.Router();
const productClient = createServiceClient(config.services.product, 'product-service');

// Public catalog reads - no auth required.
router.get('/products', async (req, res, next) => {
  try {
    const { data, status } = await productClient.get('/api/v1/products', { params: req.query });
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

router.get('/products/:id', async (req, res, next) => {
  try {
    const { data, status } = await productClient.get(`/api/v1/products/${req.params.id}`);
    res.status(status).json(data);
  } catch (err) {
    next(err);
  }
});

module.exports = router;
