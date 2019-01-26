import * as puppeteer from "puppeteer";
import * as fs from "fs";
import * as path from "path";

const done = [];

async function run() {
  const targetURL = "http://localhost:8080/";
  const browser = await puppeteer.launch();
  await crowlPage(browser, targetURL);
  await browser.close();
}

async function crowlPage(browser: puppeteer.Browser, url: string) {
  const page = await browser.newPage();
  await page.goto(url);
  await page.waitForSelector("body");
  await capture(page, ".");

  const urls = await page.$$eval("a", elms => {
    return elms.map(a => {
      return (a as HTMLAnchorElement).href;
    });
  });
  await Promise.all(
    urls
      .filter(u => !done.includes(u))
      .map(async u => {
        done.push(u);
        return await crowlPage(browser, u);
      })
  );
}

async function capture(page: puppeteer.Page, destdir: string) {
  const paths = page
    .url()
    .replace(/^https?:\/\//, "")
    .replace(/\.html$/, "")
    .replace(/\/$/, "/index")
    .split("/");

  fs.mkdirSync(path.normalize(`${destdir}/${paths.slice(0, -1).join("/")}`), {
    recursive: true
  });
  await page.screenshot({
    path: `${destdir}/${paths.join("/")}.png`,
    fullPage: true
  });
}
run();
