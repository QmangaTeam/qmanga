import { expect, test } from 'bun:test';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

for (const [workflow, job] of [['release', 'build'], ['verify', 'publish-images']]) {
  test(`${workflow} publishes lowercase image references after an organization transfer`, () => {
    const config = Bun.YAML.parse(readFileSync(join(import.meta.dir, `../.github/workflows/${workflow}.yml`), 'utf8')) as any;
    const steps = config.jobs[job].steps;
    const resolve = steps.find((step: any) => step.id === 'image');
    expect(resolve, 'workflow must resolve its image repository before publication').toBeDefined();
    const directory = mkdtempSync(join(tmpdir(), 'qmanga-image-reference-'));
    try {
      const output = join(directory, 'output');
      const result = Bun.spawnSync(['bash', '-euo', 'pipefail', '-c', resolve.run.replaceAll('${{ matrix.component }}', 'web')], {
        env: { ...process.env, GITHUB_REPOSITORY: 'QmangaTeam/Qmanga', GITHUB_OUTPUT: output },
      });
      expect(result.exitCode).toBe(0);
      expect(readFileSync(output, 'utf8').trim()).toBe('repository=ghcr.io/qmangateam/qmanga-web');
      const build = steps.find((step: any) => step.uses?.startsWith('docker/build-push-action@'));
      expect(build.with.tags).toBe('${{ steps.image.outputs.repository }}:${{ github.sha }}');
      for (const step of steps.filter((step: any) => step.run?.includes('scripts/publish-component-alias.sh'))) {
        expect(step.run).toContain('${{ steps.image.outputs.repository }}');
        expect(step.run).not.toContain('ghcr.io/${{ github.repository }}');
      }
    } finally {
      rmSync(directory, { recursive: true, force: true });
    }
  });
}
