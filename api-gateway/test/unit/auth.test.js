'use strict';

const jwt = require('jsonwebtoken');

// Isolate config per test so we can flip JWT_SECRET presence/absence -
// this exercises the "incorrect/missing environment variable" scenario.
describe('auth middleware', () => {
  const OLD_ENV = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...OLD_ENV, JWT_SECRET: 'test-secret' };
  });

  afterAll(() => {
    process.env = OLD_ENV;
  });

  function loadAuth() {
    // eslint-disable-next-line global-require
    return require('../../src/middleware/auth');
  }

  function mockRes() {
    return {
      statusCode: null,
      body: null,
      status(code) {
        this.statusCode = code;
        return this;
      },
      json(payload) {
        this.body = payload;
        return this;
      },
    };
  }

  test('rejects requests with no Authorization header', () => {
    const { authenticate } = loadAuth();
    const req = { headers: {} };
    const res = mockRes();
    const next = jest.fn();

    authenticate(req, res, next);

    expect(res.statusCode).toBe(401);
    expect(res.body.error).toBe('missing_token');
    expect(next).not.toHaveBeenCalled();
  });

  test('accepts a valid bearer token and attaches decoded user', () => {
    const { authenticate } = loadAuth();
    const token = jwt.sign({ sub: 'user-123', role: 'CUSTOMER' }, 'test-secret');
    const req = { headers: { authorization: `Bearer ${token}` } };
    const res = mockRes();
    const next = jest.fn();

    authenticate(req, res, next);

    expect(next).toHaveBeenCalled();
    expect(req.user.sub).toBe('user-123');
  });

  test('rejects an expired/invalid token', () => {
    const { authenticate } = loadAuth();
    const badToken = jwt.sign({ sub: 'user-123' }, 'wrong-secret');
    const req = { headers: { authorization: `Bearer ${badToken}` } };
    const res = mockRes();
    const next = jest.fn();

    authenticate(req, res, next);

    expect(res.statusCode).toBe(401);
    expect(res.body.error).toBe('invalid_token');
  });

  test('fails closed with a clear error when JWT_SECRET is unset', () => {
    process.env.JWT_SECRET = '';
    const { authenticate } = loadAuth();
    const req = { headers: { authorization: 'Bearer anything' } };
    const res = mockRes();
    const next = jest.fn();

    authenticate(req, res, next);

    expect(res.statusCode).toBe(500);
    expect(res.body.error).toBe('server_misconfigured');
  });
});
