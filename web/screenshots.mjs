import { chromium } from 'playwright';
import { mkdirSync, writeFileSync } from 'fs';

const BASE = 'http://localhost:8080';
const OUT = './screenshots';
mkdirSync(OUT, { recursive: true });

const pages = [
  { name: '01-home', path: '/', title: 'Home' },
  { name: '02-signin', path: '/signin', title: 'Sign In' },
  { name: '03-signup', path: '/signup', title: 'Sign Up' },
  { name: '04-watch', path: '/watch/medias-1', title: 'Watch Page' },
  { name: '05-categories', path: '/categories', title: 'Categories' },
  { name: '06-search', path: '/search', title: 'Search' },
  { name: '07-channels', path: '/channels', title: 'Channels' },
  { name: '08-settings', path: '/settings', title: 'Settings' },
  { name: '09-admin-login', path: '/admin', title: 'Admin Login' },
];

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1280, height: 800 },
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();

  const results = [];

  for (const p of pages) {
    const url = BASE + p.path;
    console.log(`Capturing: ${p.name} -> ${url}`);

    try {
      await page.goto(url, { waitUntil: 'networkidle', timeout: 15000 });
      await page.waitForTimeout(1000);
      const file = `${OUT}/${p.name}-${p.title.replace(/\s+/g, '-')}.png`;
      await page.screenshot({ path: file, fullPage: false });
      console.log(`  OK: ${file}`);
      results.push({ name: p.name, title: p.title, status: 'OK', file });
    } catch (err) {
      console.log(`  FAIL: ${err.message}`);
      results.push({ name: p.name, title: p.title, status: 'FAIL', error: err.message });
    }
  }

  const apiTests = [
    { label: 'api-health', url: `${BASE}/health` },
    { label: 'api-medias', url: `${BASE}/api/v1/medias` },
    { label: 'api-categories', url: `${BASE}/api/v1/categories` },
    { label: 'api-config', url: `${BASE}/api/v1/config` },
    { label: 'api-feed', url: `${BASE}/api/v1/feed` },
  ];

  for (const t of apiTests) {
    try {
      const resp = await page.goto(t.url, { waitUntil: 'networkidle', timeout: 10000 });
      const body = await resp.text();
      const file = `${OUT}/${t.label}.json`;
      writeFileSync(file, body.substring(0, 2000));
      console.log(`  API OK: ${t.label} ${resp.status()}`);
      results.push({ name: t.label, status: 'OK', httpStatus: resp.status() });
    } catch (err) {
      console.log(`  API FAIL: ${t.label} - ${err.message}`);
      results.push({ name: t.label, status: 'FAIL', error: err.message });
    }
  }

  await browser.close();

  console.log('\n=== Results ===');
  const pass = results.filter(r => r.status === 'OK').length;
  const fail = results.filter(r => r.status === 'FAIL').length;
  console.log(`Total: ${results.length} | Pass: ${pass} | Fail: ${fail}`);
  console.log(JSON.stringify(results, null, 2));
})();