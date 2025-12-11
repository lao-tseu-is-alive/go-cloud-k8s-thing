// Utilities
import { defineStore } from "pinia"
import { fetchAppInfo, AppInfo } from "@/tools/appInfo"
import { BACKEND_URL, getLog } from "@/config"
import { JwtCustomClaims } from "@/tools/jwt"

const log = getLog("appStore", 4, 4)
type LevelAlert = "error" | "success" | "warning" | "info"
const feedbackDefaultTimeout = 3000
export const useAppStore = defineStore("app", {
  state: () => {
    return {
      isUserAuthenticated: false,
      isNetworkOk: true,
      feedbackTimeout: feedbackDefaultTimeout, // default display time 5sec
      feedbackMsg: ` `,
      feedbackType: "success" as LevelAlert,
      feedbackVisible: false,
      appData: <AppInfo>{},
      jwtClaims: <JwtCustomClaims>{},
    }
  },
  getters: {
    getFeedbackType: (state): string => `${state.feedbackType}`,
    getFeedbackVisible: (state): boolean => state.feedbackVisible,
    getAppName: (state): string => `${state.appData.app}`,
    getAppVersion: (state): string => `${state.appData.version}`,
    getAppRepository: (state): string => `${state.appData.repository}`,
    getAppAuthUrl: (state): string => `${state.appData.authUrl}`,
    getIsUserAuthenticated: (state): boolean => state.isUserAuthenticated,
    getIsNetworkOk: (state): boolean => state.isNetworkOk,
    getUserIsAdmin: (state): boolean => {
      if (state.isUserAuthenticated) {
        const APP = state.appData.app // Use the internal appStore
        const val = sessionStorage.getItem(`${APP}_goapi_isadmin`)
        if (val !== null) {
          return val === "true"
        }
        return false
      } else {
        return false
      }
    },
    getUserId: (state): number => {
      if (state.isUserAuthenticated) {
        const APP = state.appData.app // Use the internal appStore
        const val = sessionStorage.getItem(`${APP}_goapi_idgouser`)
        if (val !== null) {
          return parseInt(val, 10)
        }
        return 0
      } else {
        return 0
      }
    },
    getUsername: (state): string => {
      const unknown = "INCONNU"
      log.t(`# entering...isUserAuthenticated:${state.isUserAuthenticated}`)
      if (state.isUserAuthenticated) {
        const APP = state.appData.app
        const val = sessionStorage.getItem(`${APP}_goapi_username`)
        if (val !== null) {
          return val
        }
        log.w(`getUsername  sessionStorage.getItem(${APP}_goapi_username is null)`)
        return unknown
      } else {
        log.w("getUsername isUserAuthenticated is false")
        return ""
      }
    },
    getUserJwtToken: (state): string => {
      log.t(`# entering...isUserAuthenticated:${state.isUserAuthenticated}`)
      if (state.isUserAuthenticated) {
        const APP = state.appData.app
        const val = sessionStorage.getItem(`${APP}_goapi_jwt_session_token`)
        if (val !== null) {
          return val
        }
        log.w(`getUserJwtToken  sessionStorage.getItem(${APP}_goapi_jwt_session_token is null)`)
        return ""
      } else {
        log.w("getUserJwtToken isUserAuthenticated is false")
        return ""
      }
    },
  },
  actions: {
    setUserAuthenticated() {
      this.isUserAuthenticated = true
    },
    setUserNotAuthenticated() {
      this.isUserAuthenticated = false
    },
    networkOffLine() {
      this.isNetworkOk = false
    },
    networkOnLine() {
      this.isNetworkOk = true
    },
    displayFeedBack(text: string, type: LevelAlert = "info", timeout: number = feedbackDefaultTimeout) {
      this.feedbackType = type
      this.feedbackMsg = text
      this.feedbackTimeout = timeout
      this.feedbackVisible = true
    },
    hideFeedBack() {
      this.feedbackVisible = false
      this.feedbackMsg = ""
      this.feedbackType = "success"
    },
    async fetchAppInfo() {
      try {
        this.appData = await fetchAppInfo(`${BACKEND_URL}/goAppInfo`)
      } catch (error) {
        this.appData = <AppInfo>{
          app: "Impossible atteindre serveur : problème réseau ?",
          version: "vous pouvez essayer de rafraîchir cette page avec la touche F5",
        }
        log.e(`Error fetching app info from ${BACKEND_URL}/goAppInfo} :`, error)
      }
    },
  },
})
