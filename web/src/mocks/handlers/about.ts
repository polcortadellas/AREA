import { http, HttpResponse } from 'msw';

import type { About } from '@/shared/api/generated/model';

export const mockAbout: About = {
  client: { host: '127.0.0.1' },
  server: {
    schema_version: 1,
    current_time: 1795217823,
    services: [
      {
        name: 'dummy',
        description: 'Test service for development: makes no external calls.',
        oauth2: false,
        actions: [
          {
            name: 'counter',
            type: 'dummy.counter',
            description: 'Fires on every poll with an increasing number.',
            inputs: [],
            outputs: [
              {
                key: 'count',
                type: 'number',
                required: true,
                description: 'Number of the event, starting at 1.',
              },
            ],
          },
        ],
        reactions: [
          {
            name: 'echo',
            type: 'dummy.echo',
            description: 'Returns the message it receives.',
            inputs: [
              {
                key: 'message',
                type: 'string',
                required: true,
                description: 'Text to return.',
              },
            ],
            outputs: [
              {
                key: 'message',
                type: 'string',
                required: true,
                description: 'The same text.',
              },
            ],
          },
        ],
      },
      {
        name: 'github',
        description: 'GitHub repositories, issues and pull requests.',
        oauth2: true,
        actions: [
          {
            name: 'new_issue',
            type: 'github.new_issue',
            description: 'A new issue is opened in a repository.',
            inputs: [
              {
                key: 'repo',
                type: 'string',
                required: true,
                description: 'Repository in owner/name format.',
              },
            ],
            outputs: [
              {
                key: 'number',
                type: 'number',
                required: true,
                description: 'Issue number.',
              },
              {
                key: 'title',
                type: 'string',
                required: true,
                description: 'Issue title.',
              },
              {
                key: 'body',
                type: 'string',
                required: false,
                description: 'Issue body text.',
              },
              {
                key: 'author',
                type: 'string',
                required: true,
                description: 'Username of the issue author.',
              },
            ],
          },
        ],
        reactions: [
          {
            name: 'create_comment',
            type: 'github.create_comment',
            description: 'Post a comment on an issue.',
            inputs: [
              {
                key: 'repo',
                type: 'string',
                required: true,
                description: 'Repository in owner/name format.',
              },
              {
                key: 'issue_number',
                type: 'number',
                required: true,
                description: 'Issue number to comment on.',
              },
              {
                key: 'body',
                type: 'string',
                required: true,
                description: 'Comment text.',
              },
            ],
            outputs: [
              {
                key: 'comment_url',
                type: 'string',
                required: true,
                description: 'URL of the created comment.',
              },
            ],
          },
        ],
      },
      {
        name: 'discord',
        description:
          'Send messages and create forum threads via Discord webhooks.',
        oauth2: false,
        actions: [],
        reactions: [
          {
            name: 'send_message',
            type: 'discord.send_message',
            description: 'Send a formatted message to a Discord webhook.',
            inputs: [
              {
                key: 'webhook_url',
                type: 'string',
                required: true,
                description: 'Discord channel webhook URL.',
              },
              {
                key: 'content',
                type: 'string',
                required: true,
                description: 'Message content.',
              },
            ],
            outputs: [
              {
                key: 'message_id',
                type: 'string',
                required: true,
                description: 'Sent message ID.',
              },
            ],
          },
          {
            name: 'create_forum_thread',
            type: 'discord.create_forum_thread',
            description: 'Create a thread in a Discord forum channel.',
            inputs: [
              {
                key: 'webhook_url',
                type: 'string',
                required: true,
                description: 'Discord forum webhook URL.',
              },
              {
                key: 'thread_name',
                type: 'string',
                required: true,
                description: 'Title of the forum thread.',
              },
              {
                key: 'content',
                type: 'string',
                required: true,
                description: 'Initial message content.',
              },
            ],
            outputs: [
              {
                key: 'thread_id',
                type: 'string',
                required: true,
                description: 'Created forum thread ID.',
              },
            ],
          },
        ],
      },
    ],
  },
};

export const aboutHandlers = [
  http.get('*/about.json', () =>
    HttpResponse.json({
      ...mockAbout,
      server: {
        ...mockAbout.server,
        current_time: Math.floor(Date.now() / 1000),
      },
    }),
  ),
];
