import { defineEventHandler, getCookie } from "h3";

export default defineEventHandler((event) => {
  const publicPage = "/authentication";
  const url = event.node.req.url || "";

  // Skip public page
  if (url.startsWith(publicPage)) return;

  // Read HttpOnly cookie
  const authToken = getCookie(event, "auth_token");

  if (!authToken) {
    return sendRedirect(event, publicPage);
  }
});
