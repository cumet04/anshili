const { captureAll } = require("capture-all");
const fs = require("fs");
const path = require("path");

const dest_dir = ".";
const urls = [
  "https://qiita.com/",
  "https://qiita.com/timeline",
  "https://qiita.com/tag-feed",
  "https://qiita.com/milestones"
];
captureAll(
  urls.map(url => {
    return {
      url: url,
      viewport: {
        width: 1280,
        height: 800
      }
    };
  })
).then(results => {
  results.forEach((result, i) => {
    const paths = result.url
      .replace(/^https?:\/\//, "")
      .replace(/\.html$/, "")
      .replace(/\/$/, "/index")
      .split("/");

    fs.mkdirSync(
      path.normalize(`${dest_dir}/${paths.slice(0, -1).join("/")}`),
      { recursive: true }
    );
    fs.writeFileSync(`${dest_dir}/${paths.join("/")}.png`, result.image);
  });
});
