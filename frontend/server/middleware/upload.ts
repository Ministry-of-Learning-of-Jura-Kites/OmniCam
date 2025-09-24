// frontend/server/middleware/upload.ts
import { defineEventHandler, sendStream } from "h3";
import { createReadStream, existsSync } from "fs";
import { join } from "path";

export default defineEventHandler(async (event) => {
  const url = event.node.req.url || "";
  const [cleanPath] = url.split("?");

  if (!cleanPath?.startsWith("/uploads/")) return;
  const relativePath = decodeURIComponent(cleanPath.replace("/uploads/", ""));
  const filePath = join(process.cwd(), "/uploads", relativePath);
  if (!existsSync(filePath)) {
    event.node.res.statusCode = 404;
    return {
      message: "File not found",
      triedPath: filePath,
      relative: relativePath,
      cwd: process.cwd(),
    };
  }
  return sendStream(event, createReadStream(filePath));
});
