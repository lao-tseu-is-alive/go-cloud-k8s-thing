// Utilities
import { defineStore } from "pinia"
import { APP, VERSION } from "@/config"

type LevelAlert = "error" | "success" | "warning" | "info"

const feedbackDefaultTimeout = 3000
export const useAppStore = defineStore("app", {
  state: () => {
    return {
      isUserAuthenticated: false,
      isNetworkOk: true,
      feedbackTimeout: feedbackDefaultTimeout, // default display time 5sec
      feedbackMsg: `${APP}, v.${VERSION}`,
      feedbackType: "success" as LevelAlert,
      feedbackVisible: false,
    }
  },
  getters: {
    getFeedbackType: (state): string => `${state.feedbackType}`,
  },
  actions: {
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
  },
})
