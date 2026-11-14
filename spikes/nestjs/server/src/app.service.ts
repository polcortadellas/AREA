import { Injectable } from '@nestjs/common';
import { PingResponseDto } from './ping.dto';

@Injectable()
export class AppService {
  ping(): PingResponseDto {
    return { status: 'ok' };
  }
}
