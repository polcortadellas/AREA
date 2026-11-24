import { describe, expect, it } from 'vitest';

import { mockAbout } from '@/mocks/handlers/about';
import { ApiError } from '@/shared/api/http';
import {
  appendKeyToken,
  buildUpstreamOutputs,
  reactionServices,
  triggerServices,
} from './catalogue';
import { extractWorkflowProblems } from './errors';
import {
  activeNodeIds,
  createDefaultReaction,
  createInitialDraft,
  fromWorkflow,
  nextReactionId,
  toWorkflowInput,
} from './workflowGraph';

describe('workflow builder catalogue & graph model', () => {
  it('filters trigger and reaction services dynamically from /about.json', () => {
    // Discord only exposes reactions in mockAbout, so it must not appear as a trigger service.
    expect(triggerServices(mockAbout).map(s => s.name)).toEqual([
      'dummy',
      'github',
    ]);
    expect(reactionServices(mockAbout).map(s => s.name)).toEqual([
      'dummy',
      'github',
      'discord',
    ]);
  });

  it('builds a linear workflow with >= 4 chained reactions and upstream key mappings', () => {
    const draft = createInitialDraft(mockAbout);
    draft.name = 'Five step echo chain';
    draft.reactions = [
      'reaction_1',
      'reaction_2',
      'reaction_3',
      'reaction_4',
    ].map((id, idx) => ({
      ...createDefaultReaction(mockAbout, id),
      config: {
        message:
          idx === 0
            ? 'Count {{ trigger.count }}'
            : `From prev {{ reaction_${idx}.message }}`,
      },
    }));

    const input = toWorkflowInput(draft, mockAbout);

    expect(input.name).toBe('Five step echo chain');
    expect(input.nodes).toHaveLength(5);
    expect(input.edges).toEqual([
      { from: 'start', to: 'reaction_1' },
      { from: 'reaction_1', to: 'reaction_2' },
      { from: 'reaction_2', to: 'reaction_3' },
      { from: 'reaction_3', to: 'reaction_4' },
    ]);

    const upstreamForFourth = buildUpstreamOutputs(
      mockAbout,
      draft.trigger.type,
      draft.reactions.slice(0, 3),
    );
    expect(upstreamForFourth.map(o => o.token)).toEqual([
      '{{ trigger.count }}',
      '{{ reaction_1.message }}',
      '{{ reaction_2.message }}',
      '{{ reaction_3.message }}',
    ]);
  });

  it('builds a branched condition workflow and round-trips through fromWorkflow', () => {
    const draft = createInitialDraft(mockAbout);
    draft.name = 'Branch on count';
    draft.condition = {
      id: 'condition',
      left: '{{ trigger.count }}',
      op: 'greater_than',
      right: '3',
    };
    draft.trueBranchReactions = [
      {
        ...createDefaultReaction(mockAbout, 'true_1'),
        config: { message: 'Big {{ trigger.count }}' },
      },
      {
        ...createDefaultReaction(mockAbout, 'true_2'),
        config: { message: 'After {{ true_1.message }}' },
      },
    ];
    draft.falseBranchReactions = [
      {
        ...createDefaultReaction(mockAbout, 'false_1'),
        config: { message: 'Small {{ trigger.count }}' },
      },
    ];

    const input = toWorkflowInput(draft, mockAbout);

    expect(input.edges).toEqual([
      { from: 'start', to: 'condition' },
      { from: 'condition', to: 'true_1', branch: 'true' },
      { from: 'true_1', to: 'true_2' },
      { from: 'condition', to: 'false_1', branch: 'false' },
    ]);

    const restored = fromWorkflow(
      {
        ...input,
        id: 'wf-1',
        enabled: true,
        created_at: '2026-11-22T10:00:00Z',
        updated_at: '2026-11-22T10:00:00Z',
      },
      mockAbout,
    );
    expect(restored.condition).toEqual(draft.condition);
    expect(restored.trueBranchReactions).toHaveLength(2);
    expect(restored.falseBranchReactions).toHaveLength(1);
    expect(nextReactionId('true', restored.trueBranchReactions)).toBe('true_3');
    expect(appendKeyToken('Alert:', '{{ trigger.count }}')).toBe(
      'Alert: {{ trigger.count }}',
    );
  });

  it('maps HTTP 422 problems to the offending step and field', () => {
    const draft = createInitialDraft(mockAbout);
    const error = new ApiError(422, {
      error: 'invalid_workflow',
      problems: [
        {
          node: 'reaction_1',
          message: '"message": node "trigger" does not declare output "nope"',
        },
        {
          message: 'a workflow has at most one condition, found 2',
        },
      ],
    });

    const parsed = extractWorkflowProblems(error, activeNodeIds(draft));

    expect(parsed.bannerKey).toBe('builder.errors.workflowProblems');
    expect(parsed.byNode.reaction_1).toEqual([
      '"message": node "trigger" does not declare output "nope"',
    ]);
    expect(parsed.byField.reaction_1?.message).toEqual([
      '"message": node "trigger" does not declare output "nope"',
    ]);
    expect(parsed.global).toEqual([
      'a workflow has at most one condition, found 2',
    ]);
  });
});
