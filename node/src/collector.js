import { mkdir, open } from 'node:fs/promises';
import { join } from 'node:path';

const API_URL = 'https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-product-by-symbol';

function ensureSymbol(symbol) {
  if (!symbol || typeof symbol !== 'string') {
    throw new Error('symbol is required');
  }
  return symbol.toUpperCase();
}

export async function collect({
  symbol,
  outputDir,
  now = new Date(),
  fetchImpl = globalThis.fetch,
} = {}) {
  if (!outputDir) {
    throw new Error('outputDir is required');
  }
  const upperSymbol = ensureSymbol(symbol);

  const url = new URL(API_URL);
  url.searchParams.set('symbol', upperSymbol);

  const response = await fetchImpl(url, {
    headers: {
      Accept: 'application/json',
      'User-Agent': 'collector-node/1.0',
    },
  });

  if (!response.ok) {
    throw new Error(`unexpected status: ${response.status}`);
  }

  const json = await response.json();
  const entry = {
    symbol: upperSymbol,
    collected_at: new Date(now).toISOString(),
    payload: json?.data ?? {},
  };

  await mkdir(outputDir, { recursive: true });
  const fileName = `${upperSymbol}-${formatHour(now)}.ndjson`;
  const filePath = join(outputDir, fileName);

  const file = await open(filePath, 'a');
  try {
    await file.appendFile(`${JSON.stringify(entry)}\n`);
  } finally {
    await file.close();
  }

  return filePath;
}

function formatHour(date) {
  const d = new Date(date);
  return [
    d.getUTCFullYear(),
    String(d.getUTCMonth() + 1).padStart(2, '0'),
    String(d.getUTCDate()).padStart(2, '0'),
    String(d.getUTCHours()).padStart(2, '0'),
  ].join('');
}

export default collect;
