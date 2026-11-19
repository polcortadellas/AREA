import { http, HttpResponse } from 'msw';

/**
 * Hand-written placeholder until openapi.yaml (M-05) lands and
 * `npm run api:gen` produces generated handlers. Shape follows the subject's /about.json.
 */
export const aboutHandlers = [
  http.get('*/about.json', () =>
    HttpResponse.json({
      client: { host: '127.0.0.1' },
      server: {
        current_time: Math.floor(Date.now() / 1000),
        services: [
          {
            name: 'github',
            actions: [
              { name: 'new_issue', description: 'A new issue is opened' },
            ],
            reactions: [
              { name: 'create_issue', description: 'Create an issue' },
            ],
          },
          {
            name: 'discord',
            actions: [],
            reactions: [
              { name: 'send_message', description: 'Send a formatted message' },
            ],
          },
        ],
      },
    }),
  ),
];
