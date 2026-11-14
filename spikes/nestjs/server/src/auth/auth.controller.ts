import { Controller, Get, Req, UseGuards } from '@nestjs/common';
import { AuthGuard } from '@nestjs/passport';
import { ApiExcludeEndpoint, ApiTags } from '@nestjs/swagger';
import type { Request } from 'express';
import type { OAuthUser } from './google.strategy';

@ApiTags('auth')
@Controller('auth')
export class AuthController {
  /** Redirects the browser to Google's consent screen. */
  @Get('google')
  @UseGuards(AuthGuard('google'))
  google(): void {}

  /** Google redirects back here with ?code=..., passport exchanges it for tokens. */
  @Get('google/callback')
  @UseGuards(AuthGuard('google'))
  @ApiExcludeEndpoint()
  callback(@Req() req: Request): { email?: string; provider: string } {
    const user = req.user as OAuthUser;
    // Spike only: a real app would persist tokens and issue its own session/JWT.
    return { provider: user.provider, email: user.email };
  }
}
