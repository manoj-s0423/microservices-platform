import express, { Express } from 'express';
import request from 'supertest';
import { createNotificationsRouter } from '../src/routes/notifications';
import { NotificationService } from '../src/services/notificationService';
import { EmailProvider } from '../src/services/emailProvider';
import { connectTestDb, clearTestDb, disconnectTestDb } from './setup';

class AlwaysDeliversProvider implements EmailProvider {
  async send() {
    return { delivered: true, providerMessageId: 'msg_test' };
  }
}

function buildApp(): Express {
  const app = express();
  app.use(express.json());
  const service = new NotificationService(new AlwaysDeliversProvider());
  app.use('/api/v1', createNotificationsRouter(service));
  return app;
}

beforeAll(async () => {
  await connectTestDb();
}, 60000);

afterEach(async () => {
  await clearTestDb();
});

afterAll(async () => {
  await disconnectTestDb();
});

describe('POST /api/v1/notifications', () => {
  const app = buildApp();

  test('creates and sends a notification, returns 201', async () => {
    const res = await request(app).post('/api/v1/notifications').send({
      userId: 'user-1',
      channel: 'EMAIL',
      type: 'ORDER_CONFIRMED',
      recipient: 'jane@example.com',
      subject: 'Order confirmed',
      body: 'Your order has shipped.',
    });

    expect(res.status).toBe(201);
    expect(res.body.status).toBe('SENT');
  });

  test('rejects a request missing required fields with 400', async () => {
    const res = await request(app).post('/api/v1/notifications').send({ userId: 'user-1' });
    expect(res.status).toBe(400);
    expect(res.body.error).toBe('validation_error');
  });
});

describe('GET /api/v1/notifications/:id', () => {
  const app = buildApp();

  test('returns 400 for a malformed id', async () => {
    const res = await request(app).get('/api/v1/notifications/not-an-object-id');
    expect(res.status).toBe(400);
  });

  test('returns 404 for a well-formed but unknown id', async () => {
    const res = await request(app).get('/api/v1/notifications/507f1f77bcf86cd799439011');
    expect(res.status).toBe(404);
  });

  test('returns the created notification by id', async () => {
    const created = await request(app).post('/api/v1/notifications').send({
      userId: 'user-1',
      channel: 'EMAIL',
      type: 'ACCOUNT_WELCOME',
      recipient: 'jane@example.com',
      body: 'Welcome!',
    });

    const res = await request(app).get(`/api/v1/notifications/${created.body._id}`);
    expect(res.status).toBe(200);
    expect(res.body.recipient).toBe('jane@example.com');
  });
});

describe('GET /api/v1/notifications', () => {
  const app = buildApp();

  test('requires a userId query param', async () => {
    const res = await request(app).get('/api/v1/notifications');
    expect(res.status).toBe(400);
  });
});
