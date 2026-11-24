import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it } from 'vitest';

import {
  getMockWorkflows,
  resetMockWorkflows,
} from '@/mocks/handlers/workflows';
import { server } from '@/mocks/node';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';

describe('DashboardPage', () => {
  beforeEach(signIn);

  it('lists workflows with status counts and passes accessibility checks', async () => {
    const { container } = renderRoute('/');

    expect(
      await screen.findByRole('heading', {
        name: 'Notify Discord on new GitHub issue',
      }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Echo counter updates' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('heading', {
        name: 'Create Discord forum thread for releases',
      }),
    ).toBeInTheDocument();

    const summary = screen.getByRole('region', { name: 'Workflow summary' });
    expect(summary).toHaveTextContent('Total3');
    expect(summary).toHaveTextContent('Active2');
    expect(summary).toHaveTextContent('Paused1');

    await expectNoA11yViolations(container);
  });

  it('filters workflows by status and search query', async () => {
    const { user } = renderRoute('/');
    await screen.findByRole('heading', {
      name: 'Notify Discord on new GitHub issue',
    });

    await user.click(
      within(screen.getByRole('group', { name: 'Filter workflows' })).getByRole(
        'button',
        { name: 'Paused' },
      ),
    );
    expect(
      screen.getByRole('heading', { name: 'Echo counter updates' }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', {
        name: 'Notify Discord on new GitHub issue',
      }),
    ).not.toBeInTheDocument();

    await user.click(
      within(screen.getByRole('group', { name: 'Filter workflows' })).getByRole(
        'button',
        { name: 'All' },
      ),
    );
    await user.type(screen.getByLabelText('Search workflows'), 'forum');
    expect(
      screen.getByRole('heading', {
        name: 'Create Discord forum thread for releases',
      }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'Echo counter updates' }),
    ).not.toBeInTheDocument();

    await user.type(screen.getByLabelText('Search workflows'), 'zzz');
    expect(screen.getByText('No matching workflows')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Clear filters' }));
    expect(
      screen.getByRole('heading', { name: 'Echo counter updates' }),
    ).toBeInTheDocument();
  });

  it('pauses an active workflow via PATCH, disables running it, and can resume it', async () => {
    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Notify Discord on new GitHub issue',
    });
    const card = heading.closest('article')!;

    const toggle = within(card).getByRole('switch', {
      name: 'Active state for Notify Discord on new GitHub issue',
    });
    const runBtn = within(card).getByRole('button', { name: 'Run now' });

    expect(toggle).toHaveAttribute('aria-checked', 'true');
    expect(runBtn).toBeEnabled();

    await user.click(toggle);

    expect(
      await screen.findByText(
        '“Notify Discord on new GitHub issue” paused. Future runs are stopped.',
      ),
    ).toBeInTheDocument();
    expect(toggle).toHaveAttribute('aria-checked', 'false');
    expect(runBtn).toBeDisabled();
    expect(
      within(card).getByText(
        'Paused — polling is stopped and no future runs will start.',
      ),
    ).toBeInTheDocument();
    expect(
      getMockWorkflows().find(
        wf => wf.name === 'Notify Discord on new GitHub issue',
      )?.enabled,
    ).toBe(false);

    await user.click(toggle);

    expect(
      await screen.findByText(
        '“Notify Discord on new GitHub issue” resumed. Future runs are enabled.',
      ),
    ).toBeInTheDocument();
    expect(toggle).toHaveAttribute('aria-checked', 'true');
    expect(runBtn).toBeEnabled();
    expect(
      getMockWorkflows().find(
        wf => wf.name === 'Notify Discord on new GitHub issue',
      )?.enabled,
    ).toBe(true);
  });

  it('rolls back and shows an alert when toggling fails', async () => {
    server.use(
      http.patch('*/workflows/:workflowId', () =>
        HttpResponse.json({ error: 'internal_error' }, { status: 500 }),
      ),
    );

    const { user } = renderRoute('/');
    const toggle = await screen.findByRole('switch', {
      name: 'Active state for Notify Discord on new GitHub issue',
    });

    await user.click(toggle);

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Something went wrong. Try again.',
    );
    expect(toggle).toHaveAttribute('aria-checked', 'true');
  });

  it('confirms deletion before calling DELETE and removes the workflow', async () => {
    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Echo counter updates',
    });
    const card = heading.closest('article')!;

    await user.click(within(card).getByRole('button', { name: 'Delete' }));
    expect(
      within(card).getByText('Delete “Echo counter updates”?'),
    ).toBeInTheDocument();

    await user.click(within(card).getByRole('button', { name: 'Cancel' }));
    expect(
      within(card).queryByText('Delete “Echo counter updates”?'),
    ).not.toBeInTheDocument();

    await user.click(within(card).getByRole('button', { name: 'Delete' }));
    await user.click(
      within(card).getByRole('button', { name: 'Delete workflow' }),
    );

    expect(
      await screen.findByText('“Echo counter updates” deleted.'),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('heading', { name: 'Echo counter updates' }),
    ).not.toBeInTheDocument();
  });

  it('shows an alert if deleting a workflow fails', async () => {
    server.use(
      http.delete('*/workflows/:workflowId', () =>
        HttpResponse.json({ error: 'not_found' }, { status: 404 }),
      ),
    );

    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Echo counter updates',
    });
    const card = heading.closest('article')!;

    await user.click(within(card).getByRole('button', { name: 'Delete' }));
    await user.click(
      within(card).getByRole('button', { name: 'Delete workflow' }),
    );

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This workflow or run could not be found.',
    );
  });

  it('triggers an active workflow run and announces confirmation', async () => {
    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Notify Discord on new GitHub issue',
    });
    const card = heading.closest('article')!;

    await user.click(within(card).getByRole('button', { name: 'Run now' }));

    expect(
      await screen.findByText(
        'Run queued for “Notify Discord on new GitHub issue”.',
      ),
    ).toBeInTheDocument();
  });

  it('shows an empty state when the user has no workflows', async () => {
    resetMockWorkflows([], []);
    renderRoute('/');

    expect(await screen.findByText('No workflows yet')).toBeInTheDocument();
    expect(
      screen.getAllByRole('link', { name: 'Create workflow' }).length,
    ).toBeGreaterThanOrEqual(1);
  });

  it('shows an error alert with retry when loading workflows fails', async () => {
    let attempts = 0;
    server.use(
      http.get('*/workflows', () => {
        attempts += 1;
        if (attempts === 1) {
          return HttpResponse.json(
            { error: 'internal_error' },
            { status: 500 },
          );
        }
        return HttpResponse.json([]);
      }),
    );

    const { user } = renderRoute('/');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not load your workflows.',
    );

    await user.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByText('No workflows yet')).toBeInTheDocument();
  });

  it('navigates from a workflow card to its execution history', async () => {
    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Notify Discord on new GitHub issue',
    });
    const card = heading.closest('article')!;

    await user.click(
      within(card).getByRole('link', { name: 'Execution history' }),
    );

    expect(
      await screen.findByRole('heading', { name: 'Run details' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('issue_trigger')).toBeInTheDocument();
  });

  it('opens the workflow builder when clicking Edit and updates the dashboard on save', async () => {
    const { user } = renderRoute('/');
    const heading = await screen.findByRole('heading', {
      name: 'Notify Discord on new GitHub issue',
    });
    const card = heading.closest('article')!;

    await user.click(within(card).getByRole('link', { name: 'Edit' }));

    const nameInput = await screen.findByLabelText('Workflow name');
    expect(nameInput).toHaveValue('Notify Discord on new GitHub issue');

    await user.clear(nameInput);
    await user.type(nameInput, 'Renamed GitHub issue alert');
    await user.click(screen.getByRole('button', { name: 'Save workflow' }));

    expect(
      await screen.findByText('Workflow validated and saved.'),
    ).toBeInTheDocument();

    await user.click(screen.getByRole('link', { name: 'Dashboard' }));
    expect(
      await screen.findByRole('heading', {
        name: 'Renamed GitHub issue alert',
      }),
    ).toBeInTheDocument();
  });
});
