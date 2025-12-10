import { levelLog, Log } from "@/log"

// eslint-disable-next-line no-undef
export const DEV = process.env.NODE_ENV === "development"
export const HOME = DEV ? "http://localhost:3000/" : "/"
// eslint-disable-next-line no-restricted-globals
const url = new URL(location.toString());
const currentURL = url.href.endsWith("/") ? url.href.slice(0, -1) : url.href;
export const GO_DEV_URL = "http://localhost:9393";
export const BACKEND_URL = DEV ? GO_DEV_URL : currentURL;
export const API_URL = "/goapi/v1";

export const getLog = (ModuleName: string, verbosityDev: levelLog, verbosityProd: levelLog) =>
  DEV ? new Log(ModuleName, verbosityDev) : new Log(ModuleName, verbosityProd)

export const defaultAxiosTimeout = 10000 // 10 sec
