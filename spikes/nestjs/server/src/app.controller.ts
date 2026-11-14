import { Controller, Get } from '@nestjs/common';
import { ApiOkResponse, ApiTags } from '@nestjs/swagger';
import { AppService } from './app.service';
import { PingResponseDto } from './ping.dto';

@ApiTags('health')
@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get('ping')
  @ApiOkResponse({ type: PingResponseDto })
  ping(): PingResponseDto {
    return this.appService.ping();
  }
}
