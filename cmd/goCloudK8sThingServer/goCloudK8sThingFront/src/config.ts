import { levelLog, Log } from "@/log"

export const APP = "goCloudK8sThing"
export const APP_TITLE = "Goéland-Thing"
export const VERSION = "0.0.7"
export const BUILD_DATE = "2024-02-22"
// eslint-disable-next-line no-undef
export const DEV = process.env.NODE_ENV === "development"
export const HOME = DEV ? "http://localhost:3000/" : "/"
// eslint-disable-next-line no-restricted-globals
const url = new URL(location.toString())
export const BACKEND_URL = DEV ? "http://localhost:9191" : url.origin
export const getLog = (ModuleName: string, verbosityDev: levelLog, verbosityProd: levelLog) =>
  DEV ? new Log(ModuleName, verbosityDev) : new Log(ModuleName, verbosityProd)

export const defaultAxiosTimeout = 10000 // 10 sec
