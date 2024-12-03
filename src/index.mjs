import {writeFile, readFile} from "node:fs/promises";
import {resolve, join} from "node:path";

import {io} from "socket.io-client";
import msgpackParser from "socket.io-msgpack-parser";

import events from "./events.mjs";
import discord from "./discord.mjs";

const CONFIG_PATH = resolve(join(import.meta.dirname, "..", "config.json"));
const config = JSON.parse(await readFile(CONFIG_PATH, "utf8"));

if (!config.token) {
  console.log("no token");
  process.exit(1);
}

try {
  const res = await fetch("http://localhost:6958");
  if (!res.ok) {
    console.log("bad response from proxy");
    process.exit(1);
  }
} catch {
  console.log("error from proxy or not running");
  process.exit(1);
}

try {
  const res = await fetch("http://localhost:6958/auth", {
    headers: {
      Cookie: `sb-wcsaaupukpdmqdjcgaoo-auth-token=${encodeURIComponent(JSON.stringify(config.token))}`,
    },
  });
  const data = await res.json();
  config.token = [data.session.access_token, data.session.refresh_token];

  await writeFile(CONFIG_PATH, JSON.stringify(config, null, "  "));
} catch (err) {
  console.log("failed to refresh token:", err);
  process.exit(1);
}

const sock = io("ws://localhost:6958", {
  autoConnect: false,
  transports: ["websocket"],
  timeout: 10000,
  auth: (login) => {
    console.log("socket auth");
    login({token: config.token[0]});
  },
  parser: msgpackParser,
  withCredentials: true,
});

sock.connect();

let presenceTimer;

sock.on(events.CONNECT, () => {
  console.log("connected");

  sock.emit(events.PRESENCE, "");

  if (presenceTimer != null) clearInterval(presenceTimer);
  setInterval(() => {
    sock.emit(events.PRESENCE, "");
  }, 30000);
});

sock.on(events.DISCONNECT, (data) => {
  console.log("disconnected:", data);
});

sock.on(events.CONNECT_ERROR, (data) => {
  console.log("failed to connect:", data.message, data.description?.message ?? data.description);

  let desc = data.description?.message;
  if (data.description != null && typeof data.description !== "object") desc = data.description.toString();

  if (data.message.toLowerCase() === "banned") {
    console.log("banned o7");
    process.exit(1);
  } else if (data.message.toLowerCase() === "no user" || data.message === "timeout") {
    process.exit(1);
  } else if (desc?.includes("403")) {
    process.exit(1);
  }
});

discord(config, sock);
