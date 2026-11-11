const puppeteer = require('../web/node_modules/puppeteer-core');
const path = require('path');
const fs = require('fs');

const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const BASE_URL = 'http://localhost:8090';

async function main() {
  const imagesDir = path.join(__dirname, '..', 'docs', 'images');
  if (!fs.existsSync(imagesDir)) {
    fs.mkdirSync(imagesDir, { recursive: true });
  }

  console.log('Launching headless Chrome via puppeteer-core...');
  const browser = await puppeteer.launch({
    executablePath: CHROME_PATH,
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu']
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });

  const tabs = [
    { id: 'dashboard', filename: 'distillforge_dashboard.png' },
    { id: 'new', filename: 'distillforge_pipeline.png' },
    { id: 'training', filename: 'distillforge_cockpit.png' },
    { id: 'datasets', filename: 'distillforge_curation.png' },
    { id: 'diff', filename: 'distillforge_diff.png' },
    { id: 'finops', filename: 'distillforge_finops.png' },
    { id: 'playground', filename: 'distillforge_playground.png' },
  ];

  for (const tab of tabs) {
    const url = `${BASE_URL}/?tab=${tab.id}`;
    console.log(`Navigating to ${url}...`);
    await page.goto(url, { waitUntil: 'networkidle0', timeout: 15000 });
    // Small buffer for charts and transitions
    await new Promise(r => setTimeout(r, 600));

    const outPath = path.join(imagesDir, tab.filename);
    await page.screenshot({ path: outPath, fullPage: false });
    console.log(`Saved screenshot: ${tab.filename}`);
  }

  // Mobile viewport captures (390x844 - iPhone 14/15)
  console.log('Capturing responsive mobile viewports...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 3, isMobile: true });

  const mobileTabs = [
    { id: 'dashboard', filename: 'distillforge_mobile_dashboard.png' },
    { id: 'finops', filename: 'distillforge_mobile_finops.png' },
  ];

  for (const mTab of mobileTabs) {
    const url = `${BASE_URL}/?tab=${mTab.id}`;
    await page.goto(url, { waitUntil: 'networkidle0', timeout: 15000 });
    await new Promise(r => setTimeout(r, 600));

    const outPath = path.join(imagesDir, mTab.filename);
    await page.screenshot({ path: outPath, fullPage: false });
    console.log(`Saved mobile screenshot: ${mTab.filename}`);
  }

  // Also capture opened mobile navigation drawer
  try {
    const menuBtn = await page.$('button[aria-label="Toggle navigation menu"]');
    if (menuBtn) {
      await menuBtn.click();
      await new Promise(r => setTimeout(r, 400));
      const menuOutPath = path.join(imagesDir, 'distillforge_mobile_menu.png');
      await page.screenshot({ path: menuOutPath, fullPage: false });
      console.log('Saved mobile menu screenshot: distillforge_mobile_menu.png');
    }
  } catch (err) {
    console.warn('Could not capture mobile menu:', err);
  }

  await browser.close();
  console.log('All real screenshots captured successfully!');
}

main().catch(err => {
  console.error('Capture failed:', err);
  process.exit(1);
});
