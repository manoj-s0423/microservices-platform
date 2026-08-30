import { Router, Request, Response, NextFunction } from 'express';
import mongoose from 'mongoose';
import { NotificationService } from '../services/notificationService';
import { createEmailProvider } from '../services/emailProvider';

export function createNotificationsRouter(service: NotificationService = new NotificationService(createEmailProvider())): Router {
  const router = Router();

  router.post('/notifications', async (req: Request, res: Response, next: NextFunction) => {
    try {
      const { userId, channel, type, recipient, subject, body } = req.body;

      if (!userId || !channel || !type || !recipient || !body) {
        return res.status(400).json({
          error: 'validation_error',
          message: 'userId, channel, type, recipient, and body are required',
        });
      }

      const notification = await service.send({ userId, channel, type, recipient, subject, body });
      const statusCode = notification.status === 'FAILED' ? 502 : 201;
      return res.status(statusCode).json(notification);
    } catch (err) {
      return next(err);
    }
  });

  router.get('/notifications/:id', async (req: Request, res: Response, next: NextFunction) => {
    try {
      if (!mongoose.Types.ObjectId.isValid(req.params.id)) {
        return res.status(400).json({ error: 'invalid_id', message: 'id must be a valid ObjectId' });
      }

      const notification = await service.getById(req.params.id);
      if (!notification) {
        return res.status(404).json({ error: 'notification_not_found', message: `Notification ${req.params.id} not found` });
      }
      return res.status(200).json(notification);
    } catch (err) {
      return next(err);
    }
  });

  router.get('/notifications', async (req: Request, res: Response, next: NextFunction) => {
    try {
      const userId = req.query.userId as string;
      if (!userId) {
        return res.status(400).json({ error: 'validation_error', message: 'userId query param is required' });
      }
      const notifications = await service.listByUser(userId);
      return res.status(200).json({ items: notifications, total: notifications.length });
    } catch (err) {
      return next(err);
    }
  });

  return router;
}
