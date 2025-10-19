#!/usr/bin/env node
'use strict';

const fsp = require('fs/promises');
const path = require('path');

const DEFAULT_SYMBOL = 'BTCUSDT';
const DEFAULT_INTERVAL_MS = 1000;
const DEFAULT_OUTPUT_DIR = path.resolve(process.cwd(), 'data', 'raw');
const DEFAULT_MAX_SAMPLES = Infinity;
const BINANCE_BASE_URL = 'https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol';

function parseArgs(argv) {
  const options = {
    symbol: DEFAULT_SYMBOL,
    intervalMs: DEFAULT_INTERVAL_MS,
    outputDir: DEFAULT_OUTPUT_DIR,
    samples: DEFAULT_MAX_SAMPLES,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (!arg.startsWith('-')) {
      continue;
    }

    switch (arg) {
      case '-s':
      case '--symbol': {
        const value = argv[i + 1];
        if (!value) {
          throw new Error(`${arg} requires a value`);
        }
        options.symbol = value.toUpperCase();
        i += 1;
        break;
      }
      case '-i':
      case '--interval': {
        const value = argv[i + 1];
        if (!value) {
          throw new Error(`${arg} requires a value`);
        }
        const parsed = Number(value);
        if (!Number.isFinite(parsed) || parsed <= 0) {
          throw new Error(`Interval must be a positive number of seconds. Received: ${value}`);
        }
        options.intervalMs = Math.round(parsed * 1000);
        i += 1;
        break;
      }
      case '-o':
      case '--output': {
        const value = argv[i + 1];
        if (!value) {
          throw new Error(`${arg} requires a value`);
        }
        options.outputDir = path.resolve(process.cwd(), value);
        i += 1;
        break;
      }
      case '-n':
      case '--samples': {
        const value = argv[i + 1];
        if (!value) {
          throw new Error(`${arg} requires a value`);
        }
        const parsed = Number(value);
        if (!Number.isInteger(parsed) || parsed <= 0) {
          throw new Error(`Samples must be a positive integer. Received: ${value}`);
        }
        options.samples = parsed;
        i += 1;
        break;
      }
      case '-h':
      case '--help':
        printHelp();
        process.exit(0);
        break;
      default:
        throw new Error(`Unknown option: ${arg}`);
    }
  }

  return options;
}

function printHelp() {
  console.log(`Usage: node bin/binance_collector.js [options]\n\n` +
    `Options:\n` +
    `  -s, --symbol <symbol>     Trading pair to poll (default: ${DEFAULT_SYMBOL})\n` +
    `  -i, --interval <seconds>  Polling interval in seconds (default: ${DEFAULT_INTERVAL_MS / 1000})\n` +
    `  -o, --output <dir>        Directory for raw JSON payloads (default: data/raw)\n` +
    `  -n, --samples <count>     Number of samples to collect before exiting (default: unlimited)\n` +
    `  -h, --help                Show this help message`);
}

function buildRequestUrl(symbol) {
  const url = new URL(BINANCE_BASE_URL);
  url.searchParams.set('symbol', symbol.toUpperCase());
  return url.toString();
}

function timestampFileName(symbol) {
  const now = new Date().toISOString().replace(/[:.]/g, '-');
  const uniqueSuffix = process.hrtime.bigint().toString();
  return `${symbol}-${now}-${uniqueSuffix}.json`;
}

async function ensureDirectory(dir) {
  await fsp.mkdir(dir, { recursive: true });
}

async function writePayload(filePath, payload) {
  await fsp.writeFile(filePath, payload);
}

async function pollOnce(options, activeControllers) {
  const { symbol, outputDir } = options;
  const url = buildRequestUrl(symbol);
  const controller = new AbortController();
  activeControllers.add(controller);
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
      signal: controller.signal,
    });

    if (!response.ok) {
      const body = await response.text();
      throw new Error(`Request failed with status ${response.status}: ${body}`);
    }

    const payload = await response.text();
    const fileName = timestampFileName(symbol);
    const filePath = path.join(outputDir, fileName);
    await writePayload(filePath, payload);
    console.log(`[${new Date().toISOString()}] Stored ${path.relative(process.cwd(), filePath)}`);
  } catch (error) {
    if (error.name === 'AbortError') {
      console.log('Fetch aborted');
    } else {
      console.error(`[${new Date().toISOString()}] Poll failed:`, error.message);
    }
  } finally {
    activeControllers.delete(controller);
  }
}

async function main() {
  let options;
  try {
    options = parseArgs(process.argv.slice(2));
  } catch (error) {
    console.error(error.message);
    printHelp();
    process.exitCode = 1;
    return;
  }

  await ensureDirectory(options.outputDir);

  let remaining = options.samples;
  const activeControllers = new Set();
  let timer = null;
  let shuttingDown = false;
  let inFlight = 0;
  let resolveShutdown;
  const shutdownPromise = new Promise((resolve) => {
    resolveShutdown = resolve;
  });

  const poll = async () => {
    if (shuttingDown) {
      return;
    }

    if (remaining !== Infinity) {
      if (remaining <= 0) {
        initiateShutdown();
        return;
      }
      remaining -= 1;
    }

    inFlight += 1;
    await pollOnce(options, activeControllers);
    inFlight -= 1;

    if (shuttingDown && inFlight === 0) {
      resolveShutdown();
    }
  };

  const initiateShutdown = () => {
    if (shuttingDown) {
      return;
    }
    shuttingDown = true;
    if (timer) {
      clearInterval(timer);
    }
    for (const controller of activeControllers) {
      controller.abort();
    }
    if (inFlight === 0) {
      resolveShutdown();
    }
  };

  const intervalMs = options.intervalMs;
  if (intervalMs < 10) {
    console.warn('Warning: intervals lower than 10ms may not be accurate in Node.js');
  }

  process.on('SIGINT', () => {
    console.log('Received SIGINT, shutting down...');
    initiateShutdown();
  });

  process.on('SIGTERM', () => {
    console.log('Received SIGTERM, shutting down...');
    initiateShutdown();
  });

  await poll();
  if (!shuttingDown) {
    timer = setInterval(() => {
      poll().catch((error) => {
        console.error('Unexpected polling error:', error);
      });
    }, intervalMs);
  }

  await shutdownPromise;
}

main().catch((error) => {
  console.error('Fatal error:', error);
  process.exitCode = 1;
});
