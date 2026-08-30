'use strict';

const request = require('supertest');
const nock = require('nock');
const createApp = require('../../src/app');
const config = require('../../src/config');

describe('GET /health', () => {
  const app = createApp();

  test('returns 200 UP without checking dependencies', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('UP');
    expect(res.body.service).toBe('api-gateway');
  });
});

describe('GET /ready', () => {
  const app = createApp();

  afterEach(() => nock.cleanAll());

  test('returns 200 when all downstream services are healthy', async () => {
    Object.values(config.services).forEach((baseURL) => {
      nock(baseURL).get('/health').reply(200, { status: 'UP' });
    });

    const res = await request(app).get('/ready');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('UP');
  });

  test('returns 503 when a downstream dependency is unavailable', async () => {
    const urls = Object.values(config.services);
    // First dependency is down; rest are healthy.
    nock(urls[0]).get('/health').replyWithError({ code: 'ECONNREFUSED' });
    urls.slice(1).forEach((baseURL) => {
      nock(baseURL).get('/health').reply(200, { status: 'UP' });
    });

    const res = await request(app).get('/ready');
    expect(res.status).toBe(503);
    expect(res.body.status).toBe('DEGRADED');
  });
});
