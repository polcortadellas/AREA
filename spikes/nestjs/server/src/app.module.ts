import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { AreasModule } from './areas/areas.module';
import { AuthModule } from './auth/auth.module';
import { HooksModule } from './hooks/hooks.module';

@Module({
  imports: [
    TypeOrmModule.forRoot({
      type: 'postgres',
      url:
        process.env.DATABASE_URL ?? 'postgres://area:area@localhost:5432/area',
      autoLoadEntities: true,
      // Spike only: auto-create tables. A real project would use migrations.
      synchronize: true,
    }),
    AuthModule,
    AreasModule,
    HooksModule,
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
