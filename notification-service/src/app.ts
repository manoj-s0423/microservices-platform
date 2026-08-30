import express, { Express, Request, Response, NextFunction } from 'express';
import helmet from 'helmet';
import cors from 'cors';
import { randomUUID } from 'crypto';
import pinoHttp from 'pino-http';

import { logger } from './logger';
import { healthRouter } from './routes/health';
import { createNotificationsRouter } from './routes/notifications';

export function createApp(): Express {
  const app = express();

  app.disable('x-powered-by');
  app.use(helmet());
  app.use(cors());
  app.use(express.json({ limit: '256kb' }));

  app.use(
    pinoHttp({
      logger,
      genReqId: (req, res) => {
        const existing = req.headers['x-request-id'];
        const id = (existing as string) || randomUUID();
        res.setHeader('x-request-id', id);
        return id;
      },
    })
  );

  app.use('/', healthRouter);
  app.use('/api/v1', createNotificationsRouter());

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
    logger.error({ err: err.message, stack: err.stack }, 'unhandled error');
    res.status(500).json({ error: 'internal_error', message: 'An unexpected error occurred' });
  });

  return app;
}
