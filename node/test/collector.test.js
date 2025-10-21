import { strict as assert } from 'node:assert';
import { mkdtemp, readFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { test } from 'node:test';

import { collect } from '../src/collector.js';

const SAMPLE_RESPONSE = {
  data: {
    symbol: 'BTCUSDT',
    price: '12345.67',
  },
};

test('collect writes ndjson with hourly rotation', async () => {
  const dir = await mkdtemp(join(tmpdir(), 'collector-node-'));
  const now = new Date(Date.UTC(2024, 2, 1, 15, 30, 0));

  const mockFetch = async () => ({
    ok: true,
    status: 200,
    async json() {
      return structuredClone(SAMPLE_RESPONSE);
    },
  });

  const filePath = await collect({
    symbol: 'btcusdt',
    outputDir: dir,
    now,
    fetchImpl: mockFetch,
  });

  const expected = join(dir, 'BTCUSDT-2024030115.ndjson');
  assert.equal(filePath, expected);

  const contents = await readFile(filePath, 'utf8');
  const lines = contents.trim().split('\n');
  assert.equal(lines.length, 1);

  const entry = JSON.parse(lines[0]);
  assert.equal(entry.symbol, 'BTCUSDT');
  assert.equal(entry.collected_at, now.toISOString());
  assert.deepEqual(entry.payload, SAMPLE_RESPONSE.data);
});
