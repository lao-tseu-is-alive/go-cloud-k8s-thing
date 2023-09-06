<template>
  <v-app>
    <v-app-bar color="primary" density="compact">
      <v-app-bar-nav-icon variant="text" @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ `${APP} v${VERSION}` }}</v-toolbar-title>
      <template v-if="isUserAuthenticated">
        <v-spacer></v-spacer>
        <v-btn variant="text" icon="mdi-magnify"></v-btn>
        <v-btn variant="text" icon="mdi-filter" title="afficher les critères de filtrage" @click="showSearchCriteria = !showSearchCriteria"></v-btn>
        <v-btn variant="text" icon="mdi-dots-vertical"></v-btn>
        <v-btn variant="text" icon="mdi-logout" title="Logout" @click="logout"></v-btn>
      </template>
    </v-app-bar>
    <v-main>
      <v-snackbar v-model="feedbackVisible" :timeout="feedbackTimeout" rounded="pill" :color="feedbackType" location="top">
        <v-alert :type="feedbackType" theme="dark" :text="feedbackMsg"></v-alert>
      </v-snackbar>
      <template v-if="isUserAuthenticated">
        <v-container>
          <template v-if="showSearchCriteria">
            <v-card density="compact" elevation="4" prepend-icon="mdi-filter">
              <template #title>
                <span class="text-h5">Critères de filtrages</span>
              </template>
              <v-card-text>
                <v-container>
                  <v-row>
                    <v-col cols="12" sm="6" md="4">
                      <v-text-field type="number"  v-model="searchLimit" density="compact" label="Limit rows" hint="The number of rows to retrieve from db" />
                    </v-col>
                    <v-col cols="12" sm="6" md="4">
                      <v-text-field v-model="searchOffset" density="compact" label="Offset row" />
                    </v-col>
                    <v-col cols="12" sm="6" md="4">
                      <v-select v-model="searchType" item-title="name" item-value="id" :items="arrListTypeThing" density="compact" label="TypeObjet*"></v-select>
                    </v-col>
                  </v-row>
                  <v-row>
                    <v-col cols="12" sm="6" md="4">
                      <v-checkbox v-model="searchInactivated" density="compact" label="Inactivated" />
                    </v-col>
                    <v-col cols="12" sm="6" md="4">
                      <v-checkbox v-model="searchValidated" density="compact" label="Validated" />
                    </v-col>
                    <v-col cols="12" sm="6" md="4">
                      <v-text-field v-model="searchCreatedBy" density="compact" label="Id of user creator" />
                    </v-col>
                  </v-row>
                </v-container>
              </v-card-text>
            </v-card>
          </template>
          <ThingList :limit="searchLimit" :offset="searchOffset" :type-thing="searchType" :created-by="searchCreatedBy" :inactivated="searchInactivated" :validated="searchValidated" />
        </v-container>
      </template>
      <template v-else>
        <Login :msg="`Authentification ${APP_TITLE}:`" :backend="APP_TITLE" :disabled="!isNetworkOk" @login-ok="loginSuccess" @login-error="loginFailure" />
      </template>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive } from "vue"
import { isNullOrUndefined } from "@/tools/utils"
import { APP, APP_TITLE, HOME, getLog, BUILD_DATE, VERSION, BACKEND_URL } from "@/config"
import Login from "@/components/Login.vue"
import ThingList from "@/components/ThingList.vue"
import { TypeThingList } from "@/typescript-axios-client-generated/models/type-thing-list"
import { getUserIsAdmin, getTokenStatus, clearSessionStorage, doesCurrentSessionExist, getLocalJwtTokenAuth } from "@/components/Login"
import { Configuration } from "@/typescript-axios-client-generated/configuration"
import { DefaultApi } from "@/typescript-axios-client-generated/apis/default-api"

const log = getLog(APP, 4, 2)
let myApi: DefaultApi
type LevelAlert = "error" | "success" | "warning" | "info" | undefined

const showSearchCriteria = ref(true)
const searchType = ref(1)
const arrListTypeThing: TypeThingList[] = reactive([])
const searchCreatedBy = ref(undefined)
const searchInactivated = ref(false)
const searchValidated = ref(undefined)
const searchLimit = ref(250)
const searchOffset = ref(0)

const isUserAuthenticated = ref(false)
const isUserAdmin = ref(false)
const isNetworkOk = ref(true)
const drawer = ref(false)
const feedbackTimeout = ref(5000) // default display time 5sec
const feedbackMsg = ref(`${APP}, v.${VERSION}`)
const feedbackType = ref()
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
  displayFeedBack("Vous êtes authentifié sur l'application.", "success")
  initialize()
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

const initialize = () => {
  log.t(`# entering...  `)
  const token = getLocalJwtTokenAuth()
  const myConf = new Configuration({ accessToken: token, basePath: BACKEND_URL + "/goapi/v1" })
  myApi = new DefaultApi(myConf)
  myApi.typeThingList(undefined, undefined, undefined, undefined, 300, 0).then((resp) => {
    log.l("myAPi.typeThingList : ", resp)
    if (resp.status == 200) {
      resp.data.forEach((r) => {
        arrListTypeThing.push(r)
      })
    } else {
      //display alert with status code > 200
    }
  })
}

onMounted(() => {
  log.t("mounted()")
  log.w(`${APP} - ${VERSION}, du ${BUILD_DATE}`)

  window.addEventListener("online", () => {
    log.w("ONLINE AGAIN :)")
    isNetworkOk.value = true
    displayFeedBack('⚡⚡🚀  LA CONNEXION RESEAU EST RÉTABLIE :  😊 vous êtes "ONLINE"  ', "success")
  })
  window.addEventListener("offline", () => {
    log.w("OFFLINE :((")
    isNetworkOk.value = false
    displayFeedBack('⚡⚡⚠ PAS DE RESEAU ! ☹ vous êtes "OFFLINE" ', "error")
  })
})
</script>
