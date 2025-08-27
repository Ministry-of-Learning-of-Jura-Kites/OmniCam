export default defineEventHandler((event) => {
  const projectId = getRouterParam(event, "projectId");

  return sendRedirect(event, `/projects/${projectId}/`, 301);
});
