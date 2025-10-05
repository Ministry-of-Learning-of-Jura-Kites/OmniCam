// scripts/proto.js
import { execSync } from "child_process";
import { platform } from "os";
import { readdir, mkdir, rm } from "node:fs/promises";
import { join } from "path";

const target = process.argv[2]; // "front", "back", or "all"
const isWin = platform() === "win32";

if (!target || !["front", "back", "all"].includes(target)) {
  console.error("Usage: node scripts/proto.js <front|back|all>");
  process.exit(1);
}

// Helper to delete folder contents (like rimraf)
async function cleanDirOrCreate(dirPath) {
  try {
    const tasks = [];
    for (const file of await readdir(dirPath)) {
      tasks.push(rm(join(dirPath, file), { recursive: true, force: true }));
    }
    await Promise.all(tasks);
    console.log(`🗑️  Cleaned ${dirPath}`);
  } catch (e) {
    if (e.code == "ENOENT") {
      await mkdir(dirPath, { recursive: true });
      console.log(`Created dir ${dirPath}`);
    } else {
      throw e;
    }
  }
}

// Helper to run a shell command
function run(cmd, label) {
  console.log(`> Running ${label}...`);
  execSync(cmd, { stdio: "inherit", shell: true });
  console.log(`✅ ${label} done`);
}

try {
  if (target === "back" || target === "all") {
    const backendDir = join("backend", "pkg", "messages", "protobufs");
    await cleanDirOrCreate(backendDir);

    const backendCmd =
      "protoc --go_out=backend/pkg/messages " +
      "--go_opt=paths=source_relative protobufs/*.proto";

    run(backendCmd, "backend proto build");
  }

  if (target === "front" || target === "all") {
    console.log();

    const frontendDir = join("frontend", "messages", "protobufs");
    await cleanDirOrCreate(frontendDir);

    const pluginPath = isWin
      ? ".\\node_modules\\.bin\\protoc-gen-ts_proto.cmd"
      : "./node_modules/.bin/protoc-gen-ts_proto";

    const frontendCmd =
      `protoc --plugin=protoc-gen-ts_proto=${pluginPath} ` +
      "--ts_proto_opt=esModuleInterop=true,forceLong=string " +
      "--ts_proto_out=frontend/messages " +
      "protobufs/*.proto";

    run(frontendCmd, "frontend proto build");
  }
} catch (err) {
  console.error("❌ Build failed:", err.message);
  process.exit(1);
}
