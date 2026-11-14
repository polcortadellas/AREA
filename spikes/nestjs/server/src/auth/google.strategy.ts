import { Injectable } from '@nestjs/common';
import { PassportStrategy } from '@nestjs/passport';
import { Profile, Strategy } from 'passport-google-oauth20';

export interface OAuthUser {
  provider: 'google';
  providerId: string;
  email?: string;
  accessToken: string;
  refreshToken?: string;
}

@Injectable()
export class GoogleStrategy extends PassportStrategy(Strategy, 'google') {
  constructor() {
    super({
      // `||` not `??`: compose passes unset vars as empty strings.
      clientID: process.env.GOOGLE_CLIENT_ID || 'missing-client-id',
      clientSecret: process.env.GOOGLE_CLIENT_SECRET || 'missing-client-secret',
      callbackURL:
        process.env.GOOGLE_CALLBACK_URL ||
        'http://localhost:8080/auth/google/callback',
      scope: ['openid', 'email', 'profile'],
      state: false,
    });
  }

  validate(
    accessToken: string,
    refreshToken: string | undefined,
    profile: Profile,
  ): OAuthUser {
    return {
      provider: 'google',
      providerId: profile.id,
      email: profile.emails?.[0]?.value,
      accessToken,
      refreshToken,
    };
  }
}
