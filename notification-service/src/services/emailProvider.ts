import { config } from '../config';
import { logger } from '../logger';

export interface SendResult {
  delivered: boolean;
  providerMessageId?: string;
  failureReason?: string;
}

export interface EmailProvider {
  send(to: string, subject: string, body: string): Promise<SendResult>;
}

/**
 * Default provider for local dev/CI/tests - no real external dependency
 * (SendGrid/SES/etc.) required. Deterministic failure so it can be
 * exercised on demand: any recipient containing "+bounce" is rejected.
 */
export class SimulatedEmailProvider implements EmailProvider {
  async send(to: string, subject: string, _body: string): Promise<SendResult> {
    await new Promise((resolve) => setTimeout(resolve, 20));

    if (to.includes('+bounce')) {
      logger.warn({ to }, 'simulated provider rejecting known-bad recipient');
      return { delivered: false, failureReason: 'recipient_bounced' };
    }

    return { delivered: true, providerMessageId: `sim_msg_${Date.now()}` };
  }
}

/**
 * "Live" provider - calls a real transactional email API (the platform's
 * external, third-party dependency for this service). Wrapped with a
 * timeout and bounded retries so a slow/unavailable provider degrades to
 * a FAILED notification instead of hanging the request indefinitely.
 */
export class HttpEmailProvider implements EmailProvider {
  async send(to: string, subject: string, body: string): Promise<SendResult> {
    let lastError: unknown;

    for (let attempt = 0; attempt <= config.emailProvider.retryAttempts; attempt++) {
      if (attempt > 0) {
        const backoff = 200 * 2 ** (attempt - 1);
        logger.warn({ attempt, backoff }, 'retrying email provider call');
        await new Promise((resolve) => setTimeout(resolve, backoff));
      }

      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), config.emailProvider.timeoutMs);

      try {
        const response = await fetch(`${config.emailProvider.apiUrl}/v1/mail/send`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${config.emailProvider.apiKey}`,
          },
          body: JSON.stringify({ to, subject, body }),
          signal: controller.signal,
        });
        clearTimeout(timeout);

        if (response.status >= 500) {
          lastError = new Error(`provider returned ${response.status}`);
          continue; // retryable
        }

        if (!response.ok) {
          const payload = (await response.json().catch(() => ({}))) as { reason?: string };
          return { delivered: false, failureReason: payload.reason || `provider_error_${response.status}` };
        }

        const payload = (await response.json()) as { messageId: string };
        return { delivered: true, providerMessageId: payload.messageId };
      } catch (err) {
        clearTimeout(timeout);
        lastError = err;
      }
    }

    logger.error({ error: lastError }, 'email provider unreachable after retries');
    return { delivered: false, failureReason: 'provider_unavailable' };
  }
}

export function createEmailProvider(): EmailProvider {
  return config.emailProvider.mode === 'live' ? new HttpEmailProvider() : new SimulatedEmailProvider();
}
