import { defineEventHandler, getCookie } from "h3";

export default defineEventHandler((event) => {
  const url = event.node.req.url || "";

  // Define routes as Regex patterns
  const publicPages = [/^\/authentication\/?$/];

  // Check if the current URL matches any of the patterns
  const isPublicPage = publicPages.some((regex) => regex.test(url));

  if (isPublicPage) return;

  const config = useRuntimeConfig();

  const authToken = getCookie(event, config.public.cookieName);

  if (!authToken) {
    return sendRedirect(event, "/authentication");
  }
});
