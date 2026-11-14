import { ValidationPipe } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger';
import helmet from 'helmet';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // Security baseline: secure headers, explicit CORS origin, strict input validation.
  app.use(helmet());
  app.enableCors({
    origin: process.env.WEB_ORIGIN ?? 'http://localhost:8081',
  });
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true, // strip unknown properties
      forbidNonWhitelisted: true, // ...and reject the request if any are sent
      transform: true,
    }),
  );

  const config = new DocumentBuilder()
    .setTitle('AREA API (NestJS spike)')
    .setVersion('0.0.1')
    .build();
  const document = SwaggerModule.createDocument(app, config);
  // UI at /docs, raw spec at /docs-json
  SwaggerModule.setup('docs', app, document, { jsonDocumentUrl: 'docs-json' });

  await app.listen(Number(process.env.PORT ?? 8080));
}
void bootstrap();
