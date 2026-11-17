const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { test } = require('node:test');
const YAML = require('yaml');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const { Manifest } = require('release-please/build/src/manifest');
const { buildStrategy } = require('release-please/build/src/factory');
const { Version } = require('release-please/build/src/version');
const { TagName } = require('release-please/build/src/util/tag-name');
const { parseConventionalCommits } = require('release-please/build/src/commit');

const root = path.resolve(__dirname, '../..');
const json = name => JSON.parse(fs.readFileSync(path.join(root, name), 'utf8'));
const workflow = name => YAML.parse(fs.readFileSync(path.join(root, '.github/workflows', name), 'utf8'));
const config = json('release-please-config.json');
const versions = json('.release-please-manifest.json');
const release = workflow('release-please.yml');
const publish = workflow('release.yml');
const logger = { debug() {}, info() {}, warn() {}, error() {} };
const github = {
  repository: { owner: 'example', repo: 'terraform-provider-example', defaultBranch: 'main' },
  getFileJson: async file => file === 'release-please-config.json' ? config : versions,
};

async function candidate(messages, previous = '0.28.0') {
  // Use release-please's actual config parser, conventional parser, versioning
  // engine and changelog updater, not a second implementation of SemVer.
  const manifest = await Manifest.fromManifest(github, 'main');
  const strategy = await buildStrategy({ ...manifest.repositoryConfig['.'], github, targetBranch: 'main', logger });
  const commits = parseConventionalCommits(messages.map((message, index) => ({
    message, sha: String(index + 1).repeat(40), files: ['internal/example.go'],
  })), logger);
  return strategy.buildReleasePullRequest(commits, {
    tag: new TagName(Version.parse(previous), undefined, undefined, true),
    sha: 'a'.repeat(40), notes: '',
  });
}

test('configuration is accepted by the pinned release-please schema', () => {
  const schema = require('release-please/schemas/config.json');
  const validate = addFormats(new Ajv({ strict: false })).compile(schema);
  assert.ok(validate(config), JSON.stringify(validate.errors));
});

for (const [message, expected] of [
  ['fix(dnsmasq): preserve explicit clears', '0.28.1'],
  ['feat(iface): add assignment controls', '0.29.0'],
  ['feat(iface)!: change identity\n\nBREAKING CHANGE: migrate the interface identity before applying.', '0.29.0'],
]) {
  test(`${message.split('\n')[0]} selects ${expected}`, async () => {
    const result = await candidate([message]);
    assert.equal(result.version.toString(), expected);
    if (message.includes('BREAKING CHANGE')) {
      assert.match(result.body.toString(), /migrate the interface identity before applying/);
    }
  });
}

test('after 1.0 a breaking change increments the major version', async () => {
  const result = await candidate(['fix!: change identity'], '1.2.3');
  assert.equal(result.version.toString(), '2.0.0');
});

test('docs/chore changes alone do not cause a provider release', async () => {
  assert.equal(await candidate(['docs: clarify setup', 'chore(ci): maintain workflows']), undefined);
});

test('parallel contributors accumulate into one version without rewriting history', async () => {
  const first = await candidate(['fix: preserve clears']);
  const combined = await candidate(['feat: add monitoring controls', 'fix: preserve clears']);
  assert.equal(first.headRefName, combined.headRefName);
  assert.equal(combined.version.toString(), '0.29.0');
  const update = combined.updates.find(item => item.path === 'CHANGELOG.md');
  const old = '# Changelog\n\n## 0.28.0 (2026-11-16)\n\nLegacy notes remain intact.\n';
  const generated = update.updater.updateContent(old, logger);
  assert.match(generated, /monitoring controls/);
  assert.match(generated, /preserve clears/);
  assert.ok(generated.endsWith(old.slice('# Changelog\n\n'.length)));
  const afterRelease = await candidate(['fix: next bug'], '0.29.0');
  assert.equal(afterRelease.version.toString(), '0.29.1');
});

test('draft release has a real tag, and signed publisher keeps its notes', async () => {
  const manifest = await Manifest.fromManifest(github, 'main');
  assert.equal(manifest.repositoryConfig['.'].forceTag, true);
  assert.equal(manifest.repositoryConfig['.'].draft, true);
  assert.equal(config.packages['.'].draft, true);
  assert.equal(config.packages['.']['force-tag-creation'], true);
  const goreleaser = YAML.parse(fs.readFileSync(path.join(root, '.goreleaser.yml'), 'utf8'));
  assert.equal(goreleaser.release.use_existing_draft, true);
  assert.equal(goreleaser.release.mode, 'keep-existing');
  assert.notEqual(goreleaser.release.draft, true); // GoReleaser publishes after uploading assets.
  assert.equal(fs.existsSync(path.join(root, '.github/workflows/auto-release-tag.yml')), false);
});

test('bot-created PR and tag both get explicit workflow dispatch', () => {
  const steps = release.jobs.release.steps;
  const action = steps.find(step => step.uses?.startsWith('googleapis/release-please-action@'));
  assert.match(action.uses, /@[a-f0-9]{40}$/);
  assert.equal(action.with.token, '${{ github.token }}');
  const pr = steps.find(step => step.name === 'Validate generated release PR');
  const tag = steps.find(step => step.name === 'Build and publish signed release assets');
  assert.match(pr.run, /gh workflow run ci.yml --ref "\$branch"/);
  assert.match(pr.if, /prs_created == 'true'/);
  assert.match(tag.run, /gh workflow run release.yml --ref "\$RELEASE_TAG"/);
  assert.match(tag.if, /release_created == 'true'/);
  assert.ok('workflow_dispatch' in workflow('ci.yml').on);
  assert.equal(release.concurrency['cancel-in-progress'], false);
  assert.match(release.jobs.release.if, /workflow_run.event == 'push'/);
});

// Execute the actual workflow shell gates with a fake read-only GitHub API.
function runGate(script, env, publication = false) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'release-gate-'));
  try {
    fs.writeFileSync(path.join(dir, 'gh'), '#!/bin/bash\ncase "$*" in\n *git/ref/heads/main*) printf "%s\\n" "$MAIN_SHA";;\n *actions/workflows/ci.yml/runs*) printf "%s\\n" "$CI_RESULT";;\n *) exit 99;;\nesac\n', { mode: 0o755 });
    if (publication) {
      fs.writeFileSync(path.join(dir, 'git'), '#!/bin/bash\ncase "$1" in\n rev-parse) printf "%s\\n" "$MAIN_SHA";;\n merge-base) exit "${ANCESTRY_EXIT:-0}";;\n *) exit 99;;\nesac\n', { mode: 0o755 });
    }
    const output = path.join(dir, 'output');
    fs.writeFileSync(output, '');
    const result = spawnSync('bash', ['-c', script], {
      env: { ...process.env, PATH: `${dir}:${process.env.PATH}`, GH_REPO: 'example/provider',
        GITHUB_OUTPUT: output, MAIN_SHA: 'a'.repeat(40), TESTED_SHA: 'a'.repeat(40),
        EVENT_NAME: 'workflow_run', CI_RESULT: 'success', ...env }, encoding: 'utf8',
    });
    return { ...result, output: fs.readFileSync(output, 'utf8') };
  } finally {
    fs.rmSync(dir, { recursive: true, force: true });
  }
}

test('stale CI is skipped; current/manual runs require successful main CI', () => {
  const gate = release.jobs.release.steps.find(step => step.id === 'gate').run;
  assert.equal(runGate(gate, {}).output, 'ready=true\n');
  const stale = runGate(gate, { TESTED_SHA: 'b'.repeat(40) });
  assert.equal(stale.status, 0);
  assert.equal(stale.output, '');
  for (const CI_RESULT of ['failure', 'cancelled', 'null']) {
    assert.notEqual(runGate(gate, { CI_RESULT }).status, 0);
    assert.notEqual(runGate(gate, { CI_RESULT, EVENT_NAME: 'workflow_dispatch' }).status, 0);
  }
  assert.equal(runGate(gate, { EVENT_NAME: 'workflow_dispatch', TESTED_SHA: '' }).output, 'ready=true\n');
});

test('publisher rejects branch refs, failed CI and non-main commits', () => {
  const gate = publish.jobs.goreleaser.steps.find(step => step.name === 'Require a main CI-verified version tag').run;
  assert.equal(runGate(gate, { GITHUB_REF: 'refs/tags/v0.29.0' }, true).status, 0);
  for (const env of [
    { GITHUB_REF: 'refs/heads/main' },
    { GITHUB_REF: 'refs/tags/v0.29.0', CI_RESULT: 'failure' },
    { GITHUB_REF: 'refs/tags/v0.29.0', ANCESTRY_EXIT: '1' },
  ]) assert.notEqual(runGate(gate, env, true).status, 0);
});
