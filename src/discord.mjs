import events from "./events.mjs";

export default function setup(config, socket) {
  if (!config.webhook) {
    console.log("not setting up discord logger, no webhook url");
    return;
  }

  async function post(data) {
    data.avatar_url = "https://www.fishtank.live/images/app/apple-icon-180-v1.png";
    data.allowed_mentions = {parse: []};

    return await fetch(config.webhook, {
      method: "POST",
      headers: {
        "User-Agent": "DiscordBot (https://hl2.zip, 1.0.0) sharknet",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
  }

  socket.on(events.NOTIFICATION_GLOBAL, (message, header) => {
    if (message.includes("A new poll") || (header && header.includes("gifted") && header.includes("season pass"))) {
      return;
    }

    post({
      username: "Fishtank : Production Message",
      content: `${header != null ? `**${header}**\n` : ""}${message}`,
    });

    console.log("Production Message:", message, header);
  });

  socket.on(events.ANNOUNCEMENT, (data) => {
    const parts = data.split(" | ");
    let colored = "\x1b[1m";

    for (const index in parts) {
      const part = parts[index];
      const i = parseInt(index) + 1;

      let col = "6";
      if (i % 3 == 0) {
        col = "3";
      } else if (i % 2 == 0) {
        col = "1";
      }

      colored += `\x1b[3${col}m${part}\x1b[39m    `;
    }

    post({
      username: "Fishtank : Ticker Updated",
      content: `\`\`\`ansi\n${colored.trim()}\n\`\`\``,
    });

    console.log("Ticker Updated:", colored.trim());
  });

  socket.on(events.POLL_START, (data) => {
    post({
      username: `Fishtank : ${data.poll.narative ? "Narative " : ""}Poll Started`,
      content: `**${data.poll.question}**\n${data.poll.answers.join("\n")}`,
    });

    console.log("Poll Started:", data);
  });
  socket.on(events.POLL_STOP, (data) => {
    post({
      username: "Fishtank : Poll Ended",
      content: `**${data.question}**\nWinner: ${data.winner}`,
    });

    console.log("Poll Ended:", data);
  });

  const features = {
    fishtoys: "Toys",
    "auto-approve-tts": "TTS Auto Approval",
    factions: "Clans",
    contestants: "Contestants",
    bigtoys: "Bigtoys",
    "live-streams": "Streams",
    quests: "Quests",
    "ai-sfx": "AI Sound Effects",
    sfx: "Sound Effects",
    tts: "TTS",
    wartoys: "Wartoys",
    "drone-cam": "Drone Cam",
  };
  socket.on(events.FEATURE_TOGGLES_UPDATED, (data) => {
    post({
      username: "Fishtank : Feature Toggle",
      content: `${features[data.feature] ?? `<unknown feature: \`${data.feature}\`>`} ${data.enabled ? "en" : "dis"}abled${data.metadata != null ? ` (metadata: \`${data.metadata}\`)` : ""}`,
    });

    console.log("Feature Toggle:", data);
  });

  socket.on(events.GOAL_CREATED, (data) => {
    post({
      username: "Fishtank : Goal Created",
      content: `**${data.name}**\n${data.description}\n\n₣${data.tokensRequired}${data.expiresAt > 0 ? `\nExpires: <t:${Math.floor(data.expiresAt / 1000)}>` : ""}`,
    });

    console.log("Goal Created:", data);
  });
  socket.on(events.GOAL_UPDATED, (data) => {
    if (data.status === "complete") {
      post({
        username: "Fishtank : Goal Updated",
        content: `Goal **${data.name}** completed.`,
      });
    }

    console.log("Goal Updated:", data);
  });
  socket.on(events.GOAL_REMOVED, (data) => {
    /*post({
      username: "Fishtank : Goal Removed",
      content: `Goal **${data.name}** removed.`,
    });*/

    console.log("Goal Removed:", data);
  });

  socket.on(events.CONTESTANT_NEW, (data) => {
    console.log("new contestant:", data);
  });
  socket.on(events.CONTESTANT_UPDATE, (data) => {
    console.log("contestant update:", data);
  });
  socket.on(events.CONTESTANT_REMOVE, (data) => {
    console.log("contestant removed:", data);
  });

  socket.on(events.HAPPENING, (data) => {
    console.log("happening:", data);
  });
  socket.on(events.PANEL_CHANGE, (data) => {
    console.log("panel change:", data);
  });
}
