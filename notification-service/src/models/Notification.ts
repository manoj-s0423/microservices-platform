import { Schema, model, Document } from 'mongoose';

export type NotificationChannel = 'EMAIL' | 'SMS';
export type NotificationStatus = 'PENDING' | 'SENT' | 'FAILED';
export type NotificationType =
  | 'ORDER_CONFIRMED'
  | 'ORDER_FAILED'
  | 'PAYMENT_RECEIPT'
  | 'ACCOUNT_WELCOME';

export interface INotification extends Document {
  userId: string;
  channel: NotificationChannel;
  type: NotificationType;
  recipient: string; // email address or phone number
  subject?: string;
  body: string;
  status: NotificationStatus;
  providerMessageId?: string;
  failureReason?: string;
  attempts: number;
  createdAt: Date;
  updatedAt: Date;
}

const notificationSchema = new Schema<INotification>(
  {
    userId: { type: String, required: true, index: true },
    channel: { type: String, enum: ['EMAIL', 'SMS'], required: true },
    type: {
      type: String,
      enum: ['ORDER_CONFIRMED', 'ORDER_FAILED', 'PAYMENT_RECEIPT', 'ACCOUNT_WELCOME'],
      required: true,
    },
    recipient: { type: String, required: true },
    subject: { type: String },
    body: { type: String, required: true },
    status: { type: String, enum: ['PENDING', 'SENT', 'FAILED'], default: 'PENDING', index: true },
    providerMessageId: { type: String },
    failureReason: { type: String },
    attempts: { type: Number, default: 0 },
  },
  { timestamps: true }
);

export const Notification = model<INotification>('Notification', notificationSchema);
