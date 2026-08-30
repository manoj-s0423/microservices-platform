'use strict';

const express = require('express');
const config = require('../config');
const { createServiceClient } = require('../services/httpClient');

const router = express.Router();

// Liveness: process is up and can serve traffic at all. Kept dependency-free
// on purpose so an orchestrator never restarts the gateway just because a
// downstream service is degraded.
router.get('/health', (req, res) => {
  res.status(200).json({ status: 'UP', service: 'api-gateway', timestamp: new Date().toISOString() });
});

// Readiness: are downstream dependencies reachable? Used by k8s readiness
// probes to pull the pod out of the Service's endpoint list without
// killing it. Intentionally short timeout per dependency so one hung
// service doesn't stall the probe past the orchestrator's own timeout.
router.get('/ready', async (req, res) => {
  const checks = await Promise.allSettled(
    Object.entries(config.services).map(async ([name, baseURL]) => {
      const client = createServiceClient(baseURL, name);
      const healthPath = config.serviceHealthPaths[name] || '/health';
      await client.get(healthPath, { timeout: 1500 });
      return name;
    })
  );

  const results = {};
  let allUp = true;
  checks.forEach((result, idx) => {
    const name = Object.keys(config.services)[idx];
    if (result.status === 'fulfilled') {
      results[name] = 'UP';
    } else {
      results[name] = 'DOWN';
      allUp = false;
    }
  });

  res.status(allUp ? 200 : 503).json({ status: allUp ? 'UP' : 'DEGRADED', dependencies: results });
});

module.exports = router;
