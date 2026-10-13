#!/usr/bin/env node
// gen_icon.js — 构建时代码生成图标（仓库不含任何图标 PNG）
//
// 产出：
//   client/assets/icon/app_icon.png            1024x1024 应用图标（flutter_launcher_icons 源图）
//   client/ios/.../LaunchImage*.png            启动图（纯品牌色，iOS 模板要求 3 个尺寸）
//
// 用法：node scripts/gen_icon.js   （build.sh / CI 在 flutter_launcher_icons 前自动调用）
const fs = require('fs');
const path = require('path');
const zlib = require('zlib');

const ROOT = path.join(__dirname, '..');
const BRAND = [12, 155, 75];      // #0C9B4B（AppDesignTokens.brand）
const WHITE = [255, 255, 255];

// ---------- PNG 编码（无依赖） ----------
const crcTable = (() => {
  const t = [];
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[n] = c >>> 0;
  }
  return t;
})();
const crc32 = (b) => {
  let c = 0xffffffff;
  for (const x of b) c = crcTable[(c ^ x) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
};
function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length);
  const t = Buffer.from(type, 'ascii');
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(Buffer.concat([t, data])));
  return Buffer.concat([len, t, data, crc]);
}
function encodePNG(w, h, rgba) {
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(w, 0);
  ihdr.writeUInt32BE(h, 4);
  ihdr[8] = 8; ihdr[9] = 6; // 8bit RGBA
  const stride = w * 4 + 1;
  const raw = Buffer.alloc(stride * h);
  for (let y = 0; y < h; y++) {
    raw[y * stride] = 0;
    rgba.copy(raw, y * stride + 1, y * w * 4, (y + 1) * w * 4);
  }
  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', ihdr),
    chunk('IDAT', zlib.deflateSync(raw, { level: 9 })),
    chunk('IEND', Buffer.alloc(0)),
  ]);
}

// ---------- 绘制 ----------
// 抗锯齿实心圆
function fillCircle(px, cx, cy, r, color, W, H) {
  for (let y = Math.max(0, Math.floor(cy - r - 1)); y < Math.min(H, cy + r + 2); y++) {
    for (let x = Math.max(0, Math.floor(cx - r - 1)); x < Math.min(W, cx + r + 2); x++) {
      const d = Math.hypot(x + 0.5 - cx, y + 0.5 - cy);
      if (d <= r + 0.5) {
        const a = Math.max(0, Math.min(1, r + 0.5 - d)); // 边缘羽化
        const p = (y * W + x) * 4;
        px[p] = Math.round(color[0] * a + px[p] * (1 - a));
        px[p + 1] = Math.round(color[1] * a + px[p + 1] * (1 - a));
        px[p + 2] = Math.round(color[2] * a + px[p + 2] * (1 - a));
      }
    }
  }
}

// 应用图标：品牌绿底 + 白色圆环 + 中心实心圆（主机指示灯意象）
function drawAppIcon(size) {
  const px = Buffer.alloc(size * size * 4);
  for (let i = 0; i < size * size; i++) {
    px[i * 4] = BRAND[0]; px[i * 4 + 1] = BRAND[1]; px[i * 4 + 2] = BRAND[2]; px[i * 4 + 3] = 255;
  }
  const c = size / 2;
  const ringR = size * 0.30, ringW = size * 0.045;
  // 外环
  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const d = Math.hypot(x + 0.5 - c, y + 0.5 - c);
      if (Math.abs(d - ringR) <= ringW / 2 + 0.5) {
        const a = Math.max(0, Math.min(1, ringW / 2 + 0.5 - Math.abs(d - ringR)));
        const p = (y * size + x) * 4;
        px[p] = Math.round(WHITE[0] * a + px[p] * (1 - a));
        px[p + 1] = Math.round(WHITE[1] * a + px[p + 1] * (1 - a));
        px[p + 2] = Math.round(WHITE[2] * a + px[p + 2] * (1 - a));
      }
    }
  }
  fillCircle(px, c, c, size * 0.14, WHITE, size, size);          // 中心点
  fillCircle(px, c + ringR, c, size * 0.055, WHITE, size, size); // 环上状态点
  return encodePNG(size, size, px);
}

// 启动图：纯品牌色（尺寸任意，Xcode 按 scale 拉伸）
function drawLaunchImage(size) {
  const px = Buffer.alloc(size * size * 4);
  for (let i = 0; i < size * size; i++) {
    px[i * 4] = BRAND[0]; px[i * 4 + 1] = BRAND[1]; px[i * 4 + 2] = BRAND[2]; px[i * 4 + 3] = 255;
  }
  return encodePNG(size, size, px);
}

// ---------- 输出 ----------
const outApp = path.join(ROOT, 'client/assets/icon/app_icon.png');
fs.mkdirSync(path.dirname(outApp), { recursive: true });
fs.writeFileSync(outApp, drawAppIcon(1024));

const launchDir = path.join(ROOT, 'client/ios/Runner/Assets.xcassets/LaunchImage.imageset');
if (fs.existsSync(launchDir)) {
  fs.writeFileSync(path.join(launchDir, 'LaunchImage.png'), drawLaunchImage(100));
  fs.writeFileSync(path.join(launchDir, 'LaunchImage@2x.png'), drawLaunchImage(200));
  fs.writeFileSync(path.join(launchDir, 'LaunchImage@3x.png'), drawLaunchImage(300));
}

console.log('✅ 图标已生成：');
console.log('   ' + outApp + ' (' + fs.statSync(outApp).size + 'B)');
if (fs.existsSync(launchDir)) console.log('   LaunchImage 1x/2x/3x 已生成');
