#!/usr/bin/env node
import collect from '../src/collector.js';

const symbol = process.argv[2] ?? 'BTCUSDT';
const outputDir = process.argv[3] ?? 'data';
const intervalArg = process.argv[4] ?? '1000';
const intervalMs = Number.parseInt(intervalArg, 10);

if (!Number.isFinite(intervalMs) || intervalMs <= 0) {
  console.error('Interval must be a positive integer representing milliseconds');
  process.exit(1);
}

let stopped = false;

process.on('SIGINT', () => {
  console.log('\nStopping collector...');
  stopped = true;
});

process.on('SIGTERM', () => {
  console.log('\nStopping collector...');
  stopped = true;
});

async function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

async function loop() {
  while (!stopped) {
    try {
      const filePath = await collect({ symbol, outputDir });
      console.log(`Snapshot stored in ${filePath}`);
    } catch (error) {
      console.error(error?.message ?? error);
    }

    if (stopped) {
      break;
    }

    await sleep(intervalMs);
  }
}

loop().catch((error) => {
  console.error(error?.message ?? error);
  process.exit(1);
});
