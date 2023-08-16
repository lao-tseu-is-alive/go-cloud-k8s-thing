import {levelLog, Log} from './log/index';

export const APP = 'GoCloudK8sThingFront';
export const APP_TITLE = 'Goéland';
export const VERSION = '0.0.5';
export const BUILD_DATE = '2023-08-16';
// eslint-disable-next-line no-undef
export const DEV = process.env.NODE_ENV === 'development';
export const HOME = DEV ? 'http://localhost:5173/' : '/';
// eslint-disable-next-line no-restricted-globals
const url = new URL(location.toString());
export const BACKEND_URL = DEV ? 'http://localhost:9191' : url.origin;
export const getLog = (ModuleName: string, verbosityDev: levelLog, verbosityProd: levelLog) => (
  (DEV) ? new Log(ModuleName, verbosityDev) : new Log(ModuleName, verbosityProd)
);
