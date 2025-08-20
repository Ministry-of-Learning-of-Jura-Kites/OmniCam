import { env } from "~/env/server";

export default defineEventHandler(() => {
  // console.info("ggg", process.env);
  // return process.env;

  return env.TEST;
});
