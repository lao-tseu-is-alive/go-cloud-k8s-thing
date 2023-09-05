<template>
  <v-app>
    <v-app-bar color="primary" prominent>
      <v-app-bar-nav-icon variant="text" @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ `${APP}:v${VERSION}` }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn variant="text" icon="mdi-magnify"></v-btn>
      <v-btn variant="text" icon="mdi-filter"></v-btn>
      <v-btn variant="text" icon="mdi-dots-vertical"></v-btn>
      <template v-if="isUserAuthenticated">
        <v-btn variant="text" icon="mdi-logout" title="Logout" @click="logout"></v-btn>
      </template>
    </v-app-bar>
    <v-main>
      <v-snackbar v-model="feedbackVisible" :timeout="feedbackTimeout"
                  rounded="pill" :color="feedbackType" location="top">
        <v-alert :type="feedbackType" :color="feedbackType" :text="feedbackMsg"> </v-alert>
      </v-snackbar>
      <template v-if="isUserAuthenticated">
        <ThingList></ThingList>
      </template>
      <template v-else>
        <Login :msg="`Authentification ${APP_TITLE}:`" :backend="APP_TITLE" :disabled="!isNetworkOk" @login-ok="loginSuccess" @login-error="loginFailure" />
      </template>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import type { Ref } from "vue"
import { isNullOrUndefined } from "@/tools/utils"
import { APP, APP_TITLE, HOME, getLog, BUILD_DATE, VERSION } from "@/config"
import Login from "@/components/Login.vue"
import ThingList from "@/components/ThingList.vue"
import { getUserIsAdmin, getTokenStatus, clearSessionStorage, doesCurrentSessionExist } from "@/components/Login"

const log = getLog(APP, 4, 2)

type LevelAlert = "success" | "info" | "warning" | "error"

const isUserAuthenticated = ref(false)
const isUserAdmin = ref(false)
const isNetworkOk = ref(true)
const drawer = ref(false)
const feedbackTimeout = ref(3000) // default display time
const feedbackMsg = ref(`${APP}, v.${VERSION}`)
const feedbackType: Ref<LevelAlert> = ref("info")
const feedbackVisible = ref(false)
let autoLogoutTimer: NodeJS.Timer | undefined
const displayFeedBack = (text: string, type: LevelAlert = "info", timeout: number = feedbackTimeout.value) => {
  log.t(`displayFeedBack() text:'${text}' type:'${type}'`)
  feedbackType.value = type
  feedbackMsg.value = text
  feedbackTimeout.value = timeout
  feedbackVisible.value = true
}

const logout = () => {
  log.t("# IN logout()")
  clearSessionStorage()
  isUserAuthenticated.value = false
  isUserAdmin.value = false
  displayFeedBack("Vous vous êtes déconnecté de l'application avec succès !", "success")
  if (isNullOrUndefined(autoLogoutTimer)) {
    clearInterval(autoLogoutTimer)
  }
  setTimeout(() => {
    window.location.href = HOME
  }, 2000) // after 2 sec redirect to home page just in case
}

const checkIsSessionTokenValid = () => {
  log.t("# entering...  ")
  if (doesCurrentSessionExist()) {
    getTokenStatus()
      .then((val) => {
        if (val.data == null) {
          log.e(`# getTokenStatus() ${val.msg}, ERROR is: `, val.err)
          displayFeedBack(`Problème réseau :${val.msg}`, "error")
        } else {
          log.l(`# getTokenStatus() SUCCESS ${val.msg} data: `, val.data)
          if (isNullOrUndefined(val.err) && val.status === 200) {
            // everything is okay, session is still valid
            isUserAuthenticated.value = true
            isUserAdmin.value = getUserIsAdmin()
            return
          }
          if (val.status === 401) {
            // jwt token is no more valid
            isUserAuthenticated.value = false
            isUserAdmin.value = false
            displayFeedBack("Votre session a expiré !", "warning")
            logout()
          }
          displayFeedBack(`Un problème est survenu avec votre session erreur: ${val.err}`, "error")
        }
      })
      .catch((err) => {
        log.e("# getJwtToken() in catch ERROR err: ", err)
        displayFeedBack(`Il semble qu'il y a eu un problème réseau ! erreur: ${err}`, "error")
      })
  } else {
    log.w("SESSION DOES NOT EXIST OR HAS EXPIRED !")
    logout()
  }
}

const loginSuccess = (v: string) => {
  log.t(`# entering... val:${v} `)
  isUserAuthenticated.value = true
  isUserAdmin.value = getUserIsAdmin()
  feedbackVisible.value = false
  displayFeedBack("Vous êtes authentifié sur l'application !", "success")
  if (isNullOrUndefined(autoLogoutTimer)) {
    // check every 60 seconds(60'000 milliseconds) if jwt is still valid
    autoLogoutTimer = setInterval(checkIsSessionTokenValid, 60000)
  }
}

const loginFailure = (v: string) => {
  log.w(`# entering... val:${v} `)
  isUserAuthenticated.value = false
  isUserAdmin.value = false
}

onMounted(() => {
  log.t("mounted()")
  log.w(`${APP} - ${VERSION}, du ${BUILD_DATE}`)

  window.addEventListener("online", () => {
    log.w("ONLINE AGAIN :)")
    isNetworkOk.value = true
    displayFeedBack('⚡⚡🚀  CONNEXION RESEAU RETABLIE :  😊 vous êtes "ONLINE"  ', "success")
  })
  window.addEventListener("offline", () => {
    log.w("OFFLINE :((")
    isNetworkOk.value = false
    displayFeedBack('⚡⚡⚠ PAS DE RESEAU ! ☹ vous êtes "OFFLINE" ', "error")
  })
})
</script>
