import { Module } from '@nestjs/common';
import { ScheduleModule } from '@nestjs/schedule';
import { AreasModule } from '../areas/areas.module';
import { HooksService } from './hooks.service';

@Module({
  imports: [ScheduleModule.forRoot(), AreasModule],
  providers: [HooksService],
})
export class HooksModule {}
