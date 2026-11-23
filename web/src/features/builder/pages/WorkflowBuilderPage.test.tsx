import { screen, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';

import { resetMockWorkflows } from '@/mocks/handlers/workflows';
import { server } from '@/mocks/node';
import type { About, WorkflowInput } from '@/shared/api/generated/model';
import { expectNoA11yViolations } from '@/test/axe';
import { renderRoute } from '@/test/render';

describe('WorkflowBuilderPage', () => {
  it('renders services, actions, reactions, inputs and output keys dynamically from /about.json with zero hardcoded services', async () => {
    const customAbout: About = {
      client: { host: '127.0.0.1' },
      server: {
        current_time: 1795217823,
        services: [
          {
            name: 'openweather',
            description: 'Live weather observations.',
            oauth2: false,
            actions: [
              {
                name: 'temp_exceeds',
                type: 'openweather.temp_exceeds',
                description: 'Temperature rises above a threshold.',
                inputs: [
                  {
                    key: 'city',
                    type: 'string',
                    required: true,
                    description: 'City name.',
                  },
                  {
                    key: 'threshold_c',
                    type: 'number',
                    required: true,
                    description: 'Threshold in Celsius.',
                  },
                ],
                outputs: [
                  {
                    key: 'temp_c',
                    type: 'number',
                    required: true,
                    description: 'Measured temperature.',
                  },
                ],
              },
            ],
            reactions: [],
          },
          {
            name: 'pagerduty',
            description: 'Incident response alerts.',
            oauth2: true,
            actions: [],
            reactions: [
              {
                name: 'trigger_incident',
                type: 'pagerduty.trigger_incident',
                description: 'Open an incident.',
                inputs: [
                  {
                    key: 'title',
                    type: 'string',
                    required: true,
                    description: 'Incident summary.',
                  },
                  {
                    key: 'urgent',
                    type: 'boolean',
                    required: false,
                    description: 'Page on-call immediately.',
                  },
                ],
                outputs: [
                  {
                    key: 'incident_id',
                    type: 'string',
                    required: true,
                    description: 'Created incident identifier.',
                  },
                ],
              },
            ],
          },
        ],
      },
    };

    server.use(http.get('*/about.json', () => HttpResponse.json(customAbout)));

    const { container } = renderRoute('/workflows/new');
    await screen.findByLabelText('Workflow name');

    const triggerStep = screen.getByTestId('step-start');
    const triggerServiceSelect = within(triggerStep).getByLabelText('Service');
    expect(
      within(triggerServiceSelect)
        .getAllByRole('option')
        .map(o => o.textContent),
    ).toEqual(['openweather']);
    expect(within(triggerStep).getByLabelText('city')).toBeInTheDocument();
    expect(
      within(triggerStep).getByLabelText('threshold_c'),
    ).toBeInTheDocument();
    expect(
      within(triggerStep).getByText('{{ trigger.temp_c }}'),
    ).toBeInTheDocument();

    const reactionStep = screen.getByTestId('step-reaction_1');
    const reactionServiceSelect =
      within(reactionStep).getByLabelText('Service');
    expect(
      within(reactionServiceSelect)
        .getAllByRole('option')
        .map(o => o.textContent),
    ).toEqual(['pagerduty']);
    expect(
      within(reactionStep).getByText('OAuth2 required'),
    ).toBeInTheDocument();
    expect(within(reactionStep).getByLabelText('title')).toBeInTheDocument();
    expect(within(reactionStep).getByLabelText('urgent')).toHaveAttribute(
      'type',
      'checkbox',
    );

    await expectNoA11yViolations(container);
  });

  it('builds and submits a demo workflow with trigger, condition, 2 branches, reactions and {{ node.key }} mapping', async () => {
    let submitted: WorkflowInput | undefined;
    server.use(
      http.post('*/workflows', async ({ request }) => {
        submitted = (await request.json()) as WorkflowInput;
        return new HttpResponse(null, { status: 200 });
      }),
    );

    const { user, container } = renderRoute('/workflows/new');
    const nameInput = await screen.findByLabelText('Workflow name');
    await user.type(nameInput, 'GitHub issue router');

    // 1. Configure Trigger: github -> new_issue
    const triggerStep = screen.getByTestId('step-start');
    await user.selectOptions(
      within(triggerStep).getByLabelText('Service'),
      'github',
    );
    await user.type(within(triggerStep).getByLabelText('repo'), 'owner/area');

    // 2. Add Condition (If / Else) and map {{ trigger.number }} > 10
    await user.click(
      screen.getByRole('button', { name: 'Add condition (If / Else)' }),
    );
    const conditionStep = screen.getByTestId('step-condition');
    await user.selectOptions(
      within(conditionStep).getByLabelText('Insert output key into left'),
      '{{ trigger.number }}',
    );
    await user.selectOptions(
      within(conditionStep).getByLabelText('Operator'),
      'greater_than',
    );
    await user.type(
      within(conditionStep).getByLabelText('Right operand'),
      '10',
    );

    // 3. Configure True branch: true_1 (dummy.echo) -> true_2 (discord.send_message)
    const trueStep1 = screen.getByTestId('step-true_1');
    await user.type(within(trueStep1).getByLabelText('message'), 'Issue');
    await user.selectOptions(
      within(trueStep1).getByLabelText('Insert output key into message'),
      '{{ trigger.title }}',
    );

    await user.click(
      screen.getByRole('button', { name: 'Add reaction to True branch' }),
    );
    const trueStep2 = screen.getByTestId('step-true_2');
    await user.selectOptions(
      within(trueStep2).getByLabelText('Service'),
      'discord',
    );
    await user.selectOptions(
      within(trueStep2).getByLabelText('Reaction'),
      'discord.send_message',
    );
    await user.type(
      within(trueStep2).getByLabelText('webhook_url'),
      'https://discord.com/api/webhooks/123/abc',
    );
    await user.selectOptions(
      within(trueStep2).getByLabelText('Insert output key into content'),
      '{{ true_1.message }}',
    );

    // 4. Configure False branch: false_1 (github.create_comment)
    const falseStep1 = screen.getByTestId('step-false_1');
    await user.selectOptions(
      within(falseStep1).getByLabelText('Service'),
      'github',
    );
    await user.type(within(falseStep1).getByLabelText('repo'), 'owner/area');
    await user.type(within(falseStep1).getByLabelText('issue_number'), '42');
    await user.selectOptions(
      within(falseStep1).getByLabelText('Insert output key into body'),
      '{{ trigger.author }}',
    );

    await expectNoA11yViolations(container);

    await user.click(screen.getByRole('button', { name: 'Save workflow' }));

    expect(await screen.findByRole('status')).toHaveTextContent(
      'Workflow validated and saved.',
    );
    expect(submitted).toEqual({
      name: 'GitHub issue router',
      enabled: true,
      nodes: [
        {
          id: 'start',
          kind: 'trigger',
          type: 'github.new_issue',
          config: { repo: 'owner/area' },
        },
        {
          id: 'condition',
          kind: 'condition',
          type: 'compare',
          config: {
            left: '{{ trigger.number }}',
            op: 'greater_than',
            right: '10',
          },
        },
        {
          id: 'true_1',
          kind: 'reaction',
          type: 'dummy.echo',
          config: { message: 'Issue {{ trigger.title }}' },
        },
        {
          id: 'true_2',
          kind: 'reaction',
          type: 'discord.send_message',
          config: {
            webhook_url: 'https://discord.com/api/webhooks/123/abc',
            content: '{{ true_1.message }}',
          },
        },
        {
          id: 'false_1',
          kind: 'reaction',
          type: 'github.create_comment',
          config: {
            repo: 'owner/area',
            issue_number: 42,
            body: '{{ trigger.author }}',
          },
        },
      ],
      edges: [
        { from: 'start', to: 'condition' },
        { from: 'condition', to: 'true_1', branch: 'true' },
        { from: 'true_1', to: 'true_2' },
        { from: 'condition', to: 'false_1', branch: 'false' },
      ],
    });
  });

  it('displays HTTP 422 validation errors inline at the offending step and marks the offending field invalid', async () => {
    const { user, container } = renderRoute('/workflows/new');
    const nameInput = await screen.findByLabelText('Workflow name');
    await user.type(nameInput, 'Broken key reference');

    // Add a second reaction and put an undeclared {{ trigger.nope }} key in reaction_2
    const reaction1 = screen.getByTestId('step-reaction_1');
    await user.selectOptions(
      within(reaction1).getByLabelText('Insert output key into message'),
      '{{ trigger.count }}',
    );

    await user.click(screen.getByRole('button', { name: 'Add reaction' }));
    const reaction2 = screen.getByTestId('step-reaction_2');
    const reaction2Input = within(reaction2).getByLabelText('message');
    // userEvent.type treats `{{` as an escaped literal `{`, so double it to type `{{`
    await user.type(reaction2Input, 'Bad {{{{ trigger.nope }}');

    await user.click(screen.getByRole('button', { name: 'Save workflow' }));

    // The HTTP 422 error appears inline inside reaction_2, not on trigger or reaction_1
    const stepAlert = await within(reaction2).findByRole('alert');
    expect(stepAlert).toHaveTextContent(
      '"message": node "trigger" does not declare output "nope"',
    );
    expect(reaction2Input).toHaveAttribute('aria-invalid', 'true');
    expect(
      within(screen.getByTestId('step-start')).queryByRole('alert'),
    ).not.toBeInTheDocument();
    expect(within(reaction1).queryByRole('alert')).not.toBeInTheDocument();

    await expectNoA11yViolations(container);
  });

  it('displays HTTP 422 empty-branch errors inline at the condition step', async () => {
    const { user } = renderRoute('/workflows/new');
    const nameInput = await screen.findByLabelText('Workflow name');
    await user.type(nameInput, 'Empty false branch');
    await user.click(
      screen.getByRole('button', { name: 'Add condition (If / Else)' }),
    );

    const conditionStep = screen.getByTestId('step-condition');
    await user.type(
      within(conditionStep).getByLabelText('Left operand'),
      '{{ trigger.count }}',
    );
    await user.type(within(conditionStep).getByLabelText('Right operand'), '5');

    const trueStep = screen.getByTestId('step-true_1');
    await user.type(within(trueStep).getByLabelText('message'), 'ok');

    // Remove the only reaction on the False branch and submit
    const falseBranch = screen.getByTestId('branch-false');
    await user.click(
      within(falseBranch).getByRole('button', { name: /Remove reaction/ }),
    );
    await user.click(screen.getByRole('button', { name: 'Save workflow' }));

    const conditionAlert = await within(conditionStep).findByRole('alert');
    expect(conditionAlert).toHaveTextContent('branch "false" is empty');
  });

  it('loads an existing workflow for editing and supports Spanish translations', async () => {
    resetMockWorkflows([
      {
        id: 'wf-existing',
        name: 'Existing counter echo',
        enabled: true,
        created_at: '2026-11-22T08:00:00Z',
        updated_at: '2026-11-22T08:00:00Z',
        nodes: [
          {
            id: 'start',
            kind: 'trigger',
            type: 'dummy.counter',
            config: {},
          },
          {
            id: 'reaction_1',
            kind: 'reaction',
            type: 'dummy.echo',
            config: { message: 'Count {{ trigger.count }}' },
          },
        ],
        edges: [{ from: 'start', to: 'reaction_1' }],
      },
    ]);

    const { user, container } = renderRoute('/workflows/wf-existing/edit');
    const nameInput = await screen.findByLabelText('Workflow name');
    expect(
      screen.getByRole('heading', { name: 'Edit workflow' }),
    ).toBeInTheDocument();
    expect(nameInput).toHaveValue('Existing counter echo');

    await user.selectOptions(screen.getByLabelText('Language'), 'es');
    expect(
      await screen.findByRole('heading', { name: 'Editar flujo de trabajo' }),
    ).toBeInTheDocument();

    await expectNoA11yViolations(container);
  });
});
