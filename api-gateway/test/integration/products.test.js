'use strict';

const request = require('supertest');
const nock = require('nock');
const createApp = require('../../src/app');
const config = require('../../src/config');

describe('GET /api/v1/products', () => {
  const app = createApp();

  afterEach(() => nock.cleanAll());

  test('proxies to product-service and returns its payload', async () => {
    nock(config.services.product)
      .get('/api/v1/products')
      .query(true)
      .reply(200, { items: [{ id: 'p1', name: 'Wireless Mouse', priceCents: 1999 }], total: 1 });

    const res = await request(app).get('/api/v1/products');

    expect(res.status).toBe(200);
    expect(res.body.items).toHaveLength(1);
  });

  test('translates a downstream timeout into 504 gateway_timeout', async () => {
    // Every retry attempt must also time out - nock interceptors are
    // single-use by default, so without `.times()` the 2nd/3rd attempt
    // would hit no mock at all and fail with a different error code.
    nock(config.services.product)
      .get('/api/v1/products')
      .query(true)
      .times(config.http.retryAttempts + 1)
      .delay(config.http.timeoutMs + 500)
      .reply(200, {});

    const res = await request(app).get('/api/v1/products');

    expect(res.status).toBe(504);
    expect(res.body.error).toBe('gateway_timeout');
  }, 15000);

  test('translates a downstream connection failure into 502 bad_gateway', async () => {
    nock(config.services.product)
      .get('/api/v1/products')
      .query(true)
      .times(3) // initial attempt + retries
      .replyWithError({ code: 'ECONNREFUSED' });

    const res = await request(app).get('/api/v1/products');

    expect(res.status).toBe(502);
    expect(res.body.error).toBe('bad_gateway');
  }, 15000);
});
