import { Injectable, Logger } from '@nestjs/common';
import { Interval } from '@nestjs/schedule';
import { AreasService } from '../areas/areas.service';

const POLL_MS = Number(process.env.HOOK_POLL_MS ?? 5000);

/**
 * Fake hook engine: every POLL_MS, check each enabled AREA's action,
 * and if it fired, run its reaction. Real actions would call external APIs
 * (in parallel, with per-service rate limits and stored cursors).
 */
@Injectable()
export class HooksService {
  private readonly logger = new Logger(HooksService.name);
  private running = false;

  constructor(private readonly areas: AreasService) {}

  @Interval(POLL_MS)
  async poll(): Promise<void> {
    if (this.running) return; // skip a tick instead of overlapping runs
    this.running = true;
    try {
      const areas = await this.areas.findEnabled();
      await Promise.allSettled(
        areas.map(async (area) => {
          // Fake action: always "fires". Fake reaction: a log line.
          this.logger.log(
            `AREA #${area.id} "${area.name}": ${area.action} -> ${area.reaction}`,
          );
          await this.areas.markTriggered(area.id);
        }),
      );
    } finally {
      this.running = false;
    }
  }
}
