import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { beforeEach, describe, expect, it } from 'vitest';

import { server } from '@/mocks/node';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute, signIn } from '@/test/render';

const GITHUB_WORKFLOW_ID = '11111111-1111-4111-8111-111111111111';
const EMPTY_RUNS_WORKFLOW_ID = '33333333-3333-4333-8333-333333333333';

describe('WorkflowHistoryPage', () => {
  beforeEach(signIn);

  it('displays runs and step-by-step status, output, skipped branch, and passes a11y', async () => {
    const { container } = renderRoute(`/workflows/${GITHUB_WORKFLOW_ID}/runs`);

    expect(
      await screen.findByRole('heading', {
        name: 'Notify Discord on new GitHub issue',
      }),
    ).toBeInTheDocument();
    expect(
      await screen.findByRole('heading', { name: 'Run details' }),
    ).toBeInTheDocument();

    // Auto-selects the newest run (SUCCEEDED) and renders every step with its output and status.
    expect(await screen.findByText('issue_trigger')).toBeInTheDocument();
    expect(screen.getByText('is_bug')).toBeInTheDocument();
    expect(screen.getByText('discord_bug_alert')).toBeInTheDocument();
    expect(screen.getByText('echo_other')).toBeInTheDocument();

    // Output JSON from trigger, condition, and reaction steps.
    expect(
      screen.getAllByText(/Fix OAuth callback redirect/).length,
    ).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/1309823412990/)).toBeInTheDocument();
    expect(screen.getByText(/#bug-alerts/)).toBeInTheDocument();

    // Skipped branch explanation.
    expect(screen.getByText('Skipped')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Skipped because this branch was not taken or an earlier step failed.',
      ),
    ).toBeInTheDocument();

    await expectNoA11yViolations(container);
  });

  it('switches to a failed run and displays the run error and failing step error', async () => {
    const { user } = renderRoute(`/workflows/${GITHUB_WORKFLOW_ID}/runs`);
    await screen.findByText('discord_bug_alert');

    await user.click(screen.getByRole('link', { name: /#aaaa2222/ }));

    const detailRegion = await screen.findByRole('region', {
      name: 'Run details',
    });
    expect(
      await within(detailRegion).findByText(
        'discord webhook returned HTTP 429: rate limited',
      ),
    ).toBeInTheDocument();
    expect(
      within(detailRegion).getByText(
        'node discord_bug_alert: discord webhook returned HTTP 429',
      ),
    ).toBeInTheDocument();
    expect(within(detailRegion).getByText('Step error')).toBeInTheDocument();
  });

  it('pauses the workflow from the history view and disables Run now', async () => {
    const { user } = renderRoute(`/workflows/${GITHUB_WORKFLOW_ID}/runs`);

    const toggle = await screen.findByRole('switch', {
      name: 'Active state for Notify Discord on new GitHub issue',
    });
    const runNow = screen.getByRole('button', { name: 'Run now' });

    expect(toggle).toHaveAttribute('aria-checked', 'true');
    expect(runNow).toBeEnabled();

    await user.click(toggle);

    expect(
      await screen.findByText(
        '“Notify Discord on new GitHub issue” paused. Future runs are stopped.',
      ),
    ).toBeInTheDocument();
    expect(toggle).toHaveAttribute('aria-checked', 'false');
    expect(runNow).toBeDisabled();
  });

  it('shows empty state for a workflow with no runs and lets the user trigger one', async () => {
    const { user } = renderRoute(`/workflows/${EMPTY_RUNS_WORKFLOW_ID}/runs`);

    expect(await screen.findByText('No runs yet')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Run now' }));

    expect(
      await screen.findByRole('heading', { name: 'Run details' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('release_trigger')).toBeInTheDocument();
    expect(screen.getByText('forum_post')).toBeInTheDocument();
  });

  it('shows a translated 404 alert when the workflow does not exist', async () => {
    renderRoute('/workflows/99999999-9999-4999-8999-999999999999/runs');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'This workflow or run could not be found.',
    );
  });

  it('shows a retryable error when run detail loading fails', async () => {
    let failedOnce = false;
    server.use(
      http.get('*/runs/:runId', () => {
        if (!failedOnce) {
          failedOnce = true;
          return HttpResponse.json(
            { error: 'internal_error' },
            { status: 500 },
          );
        }
        return HttpResponse.json({
          id: 'aaaa1111-1111-4111-8111-111111111111',
          workflow_id: GITHUB_WORKFLOW_ID,
          status: 'SUCCEEDED',
          trigger_payload: { ok: true },
          error: null,
          created_at: '2026-11-23T14:19:50.000Z',
          started_at: '2026-11-23T14:19:51.000Z',
          finished_at: '2026-11-23T14:19:52.000Z',
          steps: [
            {
              node_id: 'issue_trigger',
              status: 'SUCCEEDED',
              attempt: 1,
              output: { ok: true },
              error: null,
              started_at: '2026-11-23T14:19:51.000Z',
              finished_at: '2026-11-23T14:19:52.000Z',
            },
          ],
        });
      }),
    );

    const { user } = renderRoute(`/workflows/${GITHUB_WORKFLOW_ID}/runs`);

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Could not load details for this run.',
    );

    await user.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByText('issue_trigger')).toBeInTheDocument();
  });

  it('translates the execution history view to Spanish', async () => {
    const { user } = renderRoute(`/workflows/${GITHUB_WORKFLOW_ID}/runs`);
    await screen.findByRole('heading', { name: 'Run details' });

    await user.selectOptions(screen.getByLabelText('Language'), 'es');

    expect(
      await screen.findByRole('heading', { name: 'Detalles de la ejecución' }),
    ).toBeInTheDocument();
    expect(screen.getByText('Datos del disparador')).toBeInTheDocument();
    expect(screen.getByText('Omitido')).toBeInTheDocument();
  });
});
