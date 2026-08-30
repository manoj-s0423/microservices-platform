import { NotificationService } from '../src/services/notificationService';
import { EmailProvider, SendResult } from '../src/services/emailProvider';
import { connectTestDb, clearTestDb, disconnectTestDb } from './setup';

class FakeEmailProvider implements EmailProvider {
  constructor(private readonly result: SendResult) {}
  async send(): Promise<SendResult> {
    return this.result;
  }
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

describe('NotificationService', () => {
  test('send() persists a SENT notification when the provider delivers', async () => {
    const provider = new FakeEmailProvider({ delivered: true, providerMessageId: 'msg_1' });
    const service = new NotificationService(provider);

    const notification = await service.send({
      userId: 'user-1',
      channel: 'EMAIL',
      type: 'ORDER_CONFIRMED',
      recipient: 'jane@example.com',
      subject: 'Your order is confirmed',
      body: 'Thanks for your order!',
    });

    expect(notification.status).toBe('SENT');
    expect(notification.providerMessageId).toBe('msg_1');
    expect(notification.attempts).toBe(1);
  });

  test('send() persists a FAILED notification when the provider declines', async () => {
    const provider = new FakeEmailProvider({ delivered: false, failureReason: 'recipient_bounced' });
    const service = new NotificationService(provider);

    const notification = await service.send({
      userId: 'user-1',
      channel: 'EMAIL',
      type: 'ORDER_CONFIRMED',
      recipient: 'bad+bounce@example.com',
      body: 'Thanks for your order!',
    });

    expect(notification.status).toBe('FAILED');
    expect(notification.failureReason).toBe('recipient_bounced');
  });

  test('send() marks SMS notifications failed (channel not implemented)', async () => {
    const provider = new FakeEmailProvider({ delivered: true });
    const service = new NotificationService(provider);

    const notification = await service.send({
      userId: 'user-1',
      channel: 'SMS',
      type: 'ORDER_CONFIRMED',
      recipient: '+15555550100',
      body: 'Your order is confirmed',
    });

    expect(notification.status).toBe('FAILED');
    expect(notification.failureReason).toBe('sms_channel_not_implemented');
  });

  test('getById() returns null for an unknown but well-formed id', async () => {
    const service = new NotificationService(new FakeEmailProvider({ delivered: true }));
    const result = await service.getById('507f1f77bcf86cd799439011');
    expect(result).toBeNull();
  });

  test('listByUser() returns only that user\'s notifications, newest first', async () => {
    const provider = new FakeEmailProvider({ delivered: true, providerMessageId: 'msg' });
    const service = new NotificationService(provider);

    await service.send({ userId: 'user-a', channel: 'EMAIL', type: 'ACCOUNT_WELCOME', recipient: 'a@example.com', body: 'hi' });
    await service.send({ userId: 'user-b', channel: 'EMAIL', type: 'ACCOUNT_WELCOME', recipient: 'b@example.com', body: 'hi' });
    await service.send({ userId: 'user-a', channel: 'EMAIL', type: 'ORDER_CONFIRMED', recipient: 'a@example.com', body: 'order' });

    const results = await service.listByUser('user-a');

    expect(results).toHaveLength(2);
    expect(results.every((n) => n.userId === 'user-a')).toBe(true);
  });
});
