import { useEffect, useRef } from "react";
import "./error-pages.css";

const defaults = { 403: ["暂无访问权限", "当前账号无法访问此页面。", "forbidden"], 404: ["页面不存在", "你访问的页面不存在。", "not-found"], 500: ["服务暂时不可用", "请稍后重试。", "server"] };

function Handheld({ tone, brandName }) {
  const ref = useRef(null);
  useEffect(() => {
    const root = ref.current; const canvas = root.querySelector("canvas"); const ctx = canvas.getContext("2d"); const cols = 10; const rows = 20; const size = 14.5;
    const shapes = [[[1,1,1,1]], [[1,1],[1,1]], [[0,1,0],[1,1,1]], [[1,0,0],[1,1,1]], [[0,0,1],[1,1,1]], [[1,1,0],[0,1,1]], [[0,1,1],[1,1,0]]]; const colors = ["#696cff", "#03c3ec", "#ffab00", "#ff3e1d", "#71dd37", "#8b8dff", "#384551"];
    let board; let piece; let position; let color; let score = 0; let lines = 0; let playing = false; let timer;
    const scoreNodes = root.querySelectorAll("[data-score]"); const lineNode = root.querySelector("[data-lines]"); const startNode = root.querySelector("[data-start]");
    const block = (x, y, fill) => { ctx.fillStyle = fill; ctx.fillRect(x * size + 1, y * size + 1, size - 2, size - 2); ctx.fillStyle = "rgba(255,255,255,.22)"; ctx.fillRect(x * size + 2, y * size + 2, size - 5, 2); };
    const collides = () => piece.some((row, y) => row.some((cell, x) => cell && (position.y + y >= rows || position.x + x < 0 || position.x + x >= cols || board[position.y + y]?.[position.x + x])));
    const draw = (message = "") => { ctx.fillStyle = "#c9d7bc"; ctx.fillRect(0, 0, canvas.width, canvas.height); ctx.strokeStyle = "rgba(37,39,66,.1)"; for (let x = 0; x <= cols; x++) { ctx.beginPath(); ctx.moveTo(x * size, 0); ctx.lineTo(x * size, rows * size); ctx.stroke(); } for (let y = 0; y <= rows; y++) { ctx.beginPath(); ctx.moveTo(0, y * size); ctx.lineTo(cols * size, y * size); ctx.stroke(); } board.forEach((row, y) => row.forEach((cell, x) => cell && block(x, y, cell))); if (piece && playing) piece.forEach((row, y) => row.forEach((cell, x) => cell && block(position.x + x, position.y + y, color))); if (message) { ctx.fillStyle = "rgba(37,39,66,.78)"; ctx.fillRect(0, 0, canvas.width, rows * size); ctx.fillStyle = "#d9ff9c"; ctx.font = "700 17px monospace"; ctx.textAlign = "center"; ctx.fillText(message === "GAME OVER" ? "游戏结束" : "404 BLOCK", 100, 132); ctx.font = "700 12px monospace"; ctx.fillText(message === "GAME OVER" ? "点击开始" : message, 100, 160); } };
    const spawn = () => { piece = shapes[Math.floor(Math.random() * shapes.length)].map((row) => row.slice()); color = colors[Math.floor(Math.random() * colors.length)]; position = { x: Math.floor((cols - piece[0].length) / 2), y: 0 }; if (collides()) end(); };
    const merge = () => piece.forEach((row, y) => row.forEach((cell, x) => { if (cell) board[position.y + y][position.x + x] = color; }));
    const clearLines = () => { board = board.filter((row) => row.some((cell) => !cell)); const removed = rows - board.length; while (board.length < rows) board.unshift(Array(cols).fill(0)); score += removed * removed * 100; lines += removed; scoreNodes.forEach((node) => { node.textContent = String(score).padStart(6, "0"); }); lineNode.textContent = String(lines).padStart(2, "0"); };
    const end = () => { playing = false; clearInterval(timer); startNode.textContent = "开始"; draw("GAME OVER"); };
    const drop = () => { position.y += 1; if (collides()) { position.y -= 1; merge(); clearLines(); spawn(); } draw(); };
    const move = (delta) => { position.x += delta; if (collides()) position.x -= delta; draw(); };
    const rotate = () => { const old = piece; piece = piece[0].map((_, x) => piece.map((row) => row[x]).reverse()); if (collides()) piece = old; draw(); };
    const hardDrop = () => { while (!collides()) position.y += 1; position.y -= 1; merge(); clearLines(); spawn(); draw(); };
    const reset = () => { board = Array.from({ length: rows }, () => Array(cols).fill(0)); score = 0; lines = 0; playing = false; scoreNodes.forEach((node) => { node.textContent = "000000"; }); lineNode.textContent = "00"; draw("点击开始"); };
    const start = () => { clearInterval(timer); reset(); playing = true; startNode.textContent = "开始"; spawn(); timer = setInterval(drop, 560); draw(); };
    const action = (name) => { if (!playing) return; if (name === "left") move(-1); if (name === "right") move(1); if (name === "down") drop(); if (name === "rotate") rotate(); if (name === "drop") hardDrop(); };
    const keydown = (event) => { const key = event.key.toLowerCase(); const code = event.code.toLowerCase(); const map = { arrowleft: "left", keya: "left", a: "left", arrowright: "right", keyd: "right", d: "right", arrowdown: "down", keys: "down", s: "down", arrowup: "rotate", keyw: "rotate", w: "rotate", space: "drop" }; const command = map[key] || map[code]; if (command) { event.preventDefault(); action(command); } };
    startNode.addEventListener("click", start); root.querySelectorAll("[data-action]").forEach((button) => button.addEventListener("click", () => action(button.dataset.action))); document.addEventListener("keydown", keydown); reset();
    return () => { clearInterval(timer); startNode.removeEventListener("click", start); document.removeEventListener("keydown", keydown); };
  }, []);
  return <section ref={ref} className={`global-handheld global-handheld-${tone}`} aria-label="掌机错误页面"><div className="global-handheld-brand"><span>{brandName || "…"}</span></div><div className="global-handheld-screen-wrap"><canvas width="300" height="290" aria-label="俄罗斯方块游戏画面" /><div className="global-handheld-data"><div><span>分数 SCORE</span><strong data-score>000000</strong></div><div><span>等级 LEVEL</span><strong>01</strong></div><div><span>行数 LINES</span><strong data-lines>00</strong></div></div></div><div className="global-handheld-score"><span>SCORE <b data-score>000000</b></span><span>LEVEL 01</span></div><div className="global-handheld-controls"><div className="global-dpad"><button data-action="up" aria-label="旋转">▲</button><button data-action="left" aria-label="左移">◀</button><button data-action="down" aria-label="下移">▼</button><button data-action="right" aria-label="右移">▶</button></div><div className="global-actions"><button data-action="rotate" aria-label="旋转方块">A</button><button data-action="drop" aria-label="快速下落">B</button><button data-action="rotate" aria-label="旋转方块">X</button><button data-action="drop" aria-label="快速下落">Y</button></div></div><div className="global-system-buttons"><button type="button" aria-label="选择">选择</button><button data-start className="global-start" type="button">开始</button></div></section>;
}

export function GlobalErrorPage({ code = 404, title, description, action, onAction, brandName }) {
  const [defaultTitle, defaultDescription, tone] = defaults[code] || defaults[404];
  return <main className={`global-error-page global-error-${tone}`} aria-labelledby={`global-error-title-${code}`}><h1 className="global-error-number" id={`global-error-title-${code}`}>{code}</h1><p className="global-error-sr-only">{title || defaultTitle}。{description || defaultDescription}</p><Handheld tone={tone} brandName={brandName} /></main>;
}
