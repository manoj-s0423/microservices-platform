import request from 'supertest';
import { createApp } from '../src/app';
import { connectTestDb, disconnectTestDb } from './setup';

describe('health endpoints', () => {
  const app = createApp();

  test('GET /health returns 200 UP without checking dependencies', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('UP');
    expect(res.body.service).toBe('notification-service');
  });

  test('GET /ready returns 503 when MongoDB is not connected', async () => {
    const res = await request(app).get('/ready');
    expect(res.status).toBe(503);
    expect(res.body.dependencies.mongodb).toBe('DOWN');
  });

  describe('when connected', () => {
    beforeAll(async () => {
      await connectTestDb();
    }, 60000);

    afterAll(async () => {
      await disconnectTestDb();
    });

    test('GET /ready returns 200 when MongoDB is connected', async () => {
      const res = await request(app).get('/ready');
      expect(res.status).toBe(200);
      expect(res.body.dependencies.mongodb).toBe('UP');
    });
  });
});
