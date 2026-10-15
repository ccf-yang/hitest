// 生成占位 AppIcon: 1024x1024 纯色圆角风格 PNG (无第三方依赖, 纯 zlib 手工编码)。
// 用途: flutter_launcher_icons 的源图 (pubspec.yaml → assets/icon/app_icon.png)。
// 替换正式图标时直接覆盖该文件即可。
const zlib = require('zlib');
const fs = require('fs');
const path = require('path');

const SIZE = 1024;

// 深蓝底 + 居中浅色方块 (模拟 logo)
function pixel(x, y) {
  // 背景: #16181d
  let r = 0x16, g = 0x18, b = 0x1d;
  // 中央圆角方块: #4f9cf9
  const cx = SIZE / 2, cy = SIZE / 2, half = 280, radius = 80;
  const dx = Math.abs(x - cx), dy = Math.abs(y - cy);
  const inBox = dx < half && dy < half;
  const corner =
    dx > half - radius && dy > half - radius &&
    Math.hypot(dx - (half - radius), dy - (half - radius)) > radius;
  if (inBox && !corner) {
    r = 0x4f; g = 0x9c; b = 0xf9;
  }
  return [r, g, b];
}

// PNG 编码
function crc32(buf) {
  let table = crc32.table;
  if (!table) {
    table = crc32.table = [];
    for (let n = 0; n < 256; n++) {
      let c = n;
      for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
      table[n] = c >>> 0;
    }
  }
  let c = 0xffffffff;
  for (const byte of buf) c = table[(c ^ byte) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}

function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length);
  const body = Buffer.concat([Buffer.from(type), data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(body));
  return Buffer.concat([len, body, crc]);
}

const ihdr = Buffer.alloc(13);
ihdr.writeUInt32BE(SIZE, 0);
ihdr.writeUInt32BE(SIZE, 4);
ihdr[8] = 8;  // bit depth
ihdr[9] = 2;  // color type: truecolor RGB (无 alpha, 符合 remove_alpha_ios)

// 原始扫描线: 每行前缀 filter byte 0
const raw = Buffer.alloc(SIZE * (1 + SIZE * 3));
let off = 0;
for (let y = 0; y < SIZE; y++) {
  raw[off++] = 0;
  for (let x = 0; x < SIZE; x++) {
    const [r, g, b] = pixel(x, y);
    raw[off++] = r; raw[off++] = g; raw[off++] = b;
  }
}

const png = Buffer.concat([
  Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
  chunk('IHDR', ihdr),
  chunk('IDAT', zlib.deflateSync(raw, { level: 9 })),
  chunk('IEND', Buffer.alloc(0)),
]);

const out = path.join(__dirname, '..', 'client', 'assets', 'icon', 'app_icon.png');
fs.mkdirSync(path.dirname(out), { recursive: true });
fs.writeFileSync(out, png);
console.log(`生成 ${out} (${png.length} bytes, ${SIZE}x${SIZE})`);
