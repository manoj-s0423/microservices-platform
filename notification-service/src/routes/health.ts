import { Router, Request, Response } from 'express';
import { isDatabaseConnected } from '../db';

export const healthRouter = Router();

// Liveness: process is up. No dependency checks on purpose.
healthRouter.get('/health', (_req: Request, res: Response) => {
  res.status(200).json({ status: 'UP', service: 'notification-service', timestamp: new Date().toISOString() });
});

// Readiness: can we reach MongoDB?
healthRouter.get('/ready', (_req: Request, res: Response) => {
  const dbUp = isDatabaseConnected();
  res.status(dbUp ? 200 : 503).json({
    status: dbUp ? 'UP' : 'DEGRADED',
    dependencies: { mongodb: dbUp ? 'UP' : 'DOWN' },
  });
});
