import { Notification, INotification, NotificationChannel, NotificationType } from '../models/Notification';
import { EmailProvider } from './emailProvider';
import { logger } from '../logger';

export interface SendNotificationRequest {
  userId: string;
  channel: NotificationChannel;
  type: NotificationType;
  recipient: string;
  subject?: string;
  body: string;
}

export class NotificationService {
  constructor(private readonly emailProvider: EmailProvider) {}

  async send(request: SendNotificationRequest): Promise<INotification> {
    const notification = new Notification({
      userId: request.userId,
      channel: request.channel,
      type: request.type,
      recipient: request.recipient,
      subject: request.subject,
      body: request.body,
      status: 'PENDING',
      attempts: 0,
    });

    await notification.save();

    if (request.channel === 'SMS') {
      // SMS provider integration is intentionally out of scope for this
      // portfolio build - the schema/route/model support it end-to-end,
      // but only EMAIL is actually dispatched. Mark PENDING notifications
      // for SMS as failed with a clear, honest reason rather than
      // pretending to send.
      notification.status = 'FAILED';
      notification.failureReason = 'sms_channel_not_implemented';
      notification.attempts = 1;
      await notification.save();
      return notification;
    }

    notification.attempts += 1;
    const result = await this.emailProvider.send(
      notification.recipient,
      notification.subject || '',
      notification.body
    );

    if (result.delivered) {
      notification.status = 'SENT';
      notification.providerMessageId = result.providerMessageId;
    } else {
      notification.status = 'FAILED';
      notification.failureReason = result.failureReason;
      logger.warn({ notificationId: notification.id, reason: result.failureReason }, 'notification failed to send');
    }

    await notification.save();
    return notification;
  }

  async getById(id: string): Promise<INotification | null> {
    return Notification.findById(id);
  }

  async listByUser(userId: string): Promise<INotification[]> {
    return Notification.find({ userId }).sort({ createdAt: -1 });
  }
}
