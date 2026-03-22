import "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    // Define the structure of routeInfo here
    routeInfo?: {
      workspace?: string;
      // add other properties as needed
    };
  }
}
