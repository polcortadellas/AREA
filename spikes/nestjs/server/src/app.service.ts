import { Injectable } from '@nestjs/common';

@Injectable()
export class AppService {
  ping(): { status: string } {
    return { status: 'ok' };
  }
}
