import "~/env/server";

export default defineNitroPlugin(() => {
  console.log("Environment validated at runtime");
});
