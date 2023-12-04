<template>
  <v-app>
    <v-app-bar color="primary" density="compact">
      <v-app-bar-nav-icon variant="text" @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ `${APP} v${VERSION}` }}</v-toolbar-title>
      <template v-if="DEV"
        ><span class="left-0">{{ displaySize.name }}</span></template
      >
      <template v-if="isUserAuthenticated">
        <v-spacer></v-spacer>
        <v-btn variant="text" :icon="getMapIcon()" :title="getMapTitle()" @click="showMap = !showMap"> </v-btn>
        <v-btn variant="text" icon="mdi-magnify"></v-btn>
        <v-btn
          variant="text"
          :icon="getFilterIcon()"
          :title="getFilterTitle()"
          @click="showSearchCriteria = !showSearchCriteria"
        ></v-btn>
        <v-btn variant="text" icon="mdi-dots-vertical" title="Réglages" @click="showSettings = !showSettings"></v-btn>
        <v-btn variant="text" icon="mdi-logout" title="Logout" @click="logout"></v-btn>
      </template>
    </v-app-bar>
    <v-main>
      <v-snackbar
        v-model="feedbackVisible"
        :timeout="feedbackTimeout"
        rounded="pill"
        :color="feedbackType"
        location="top"
      >
        <v-alert class="ma-4" :type="feedbackType" :text="feedbackMsg" :color="feedbackType"></v-alert>
      </v-snackbar>
      <template v-if="isUserAuthenticated">
        <v-responsive class="text-center pa-2">
          <template v-if="showSearchCriteria && !showSettings">
            <v-row>
              <v-col cols="12">
                <v-card density="compact" elevation="4" prepend-icon="mdi-filter">
                  <template #title>
                    <span class="text-h5">Critères de filtrage</span>
                  </template>
                  <v-card-text>
                    <v-container>
                      <v-row>
                        <v-col cols="12" sm="6" md="6" lg="2" xl="2">
                          <v-text-field
                            v-model="searchCreatedBy"
                            type="number"
                            min="1"
                            density="compact"
                            title="id de l'utilisateur qui a créé l'enregistrement"
                            label="id créateur enregistrement"
                          />
                        </v-col>
                        <v-col cols="12" sm="6" md="6" lg="2" xl="2">
                          <v-select
                            v-model="searchType"
                            item-title="name"
                            item-value="id"
                            :items="arrListTypeThing"
                            density="compact"
                            label="TypeObjet*"
                          ></v-select>
                        </v-col>
                        <v-col cols="6" sm="3" md="3" lg="2" xl="1">
                          <v-checkbox v-model="searchInactivated" density="compact" label="Inactivé ?" />
                        </v-col>
                        <v-col cols="6" sm="3" md="3" lg="2" xl="1">
                          <v-checkbox v-model="searchValidated" density="compact" label="Validé ? " />
                        </v-col>
                        <v-col cols="12" sm="6" md="6" lg="4" xl="6">
                          <v-text-field v-model="searchKeywords" density="compact" label="mot clés" />
                        </v-col>
                      </v-row>
                    </v-container>
                  </v-card-text>
                  <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn dark color="primary" variant="flat" prepend-icon="mdi-eraser" @click.prevent="clearFilters"
                      >Réinitialiser Filtres</v-btn
                    >
                  </v-card-actions>
                </v-card>
              </v-col>
            </v-row>
          </template>
          <template v-else-if="showSettings">
            <v-card density="compact" elevation="4" prepend-icon="mdi-filter">
              <template #title>
                <span class="text-h5">Réglages</span>
              </template>
              <v-card-text>
                <v-container>
                  <v-row>
                    <v-col cols="12" sm="6" md="4" lg="4" xl="4">
                      <v-text-field
                        type="number"
                        :rules="[rules.required, rules.minNumber1]"
                        min="1"
                        v-model="searchLimit"
                        density="compact"
                        label="Max rows"
                        hint="Le nombre maximum d'enregistrements à récupérer dans la Base de données"
                      />
                    </v-col>
                    <v-col cols="12" sm="6" md="4" lg="4" xl="4">
                      <v-text-field type="number" v-model="searchOffset" density="compact" min="0" label="Offset row" />
                    </v-col>
                  </v-row>
                </v-container>
              </v-card-text>
            </v-card>
          </template>
        </v-responsive>
        <v-container class="text-center fill-height pa-2" :fluid="true">
          <v-row class="fill-height h-100">
            <v-col cols="12" v-show="showMap" class="fill-height">
              <MapLausanne :zoom="3" @map-click="mapClickHandler"></MapLausanne>
            </v-col>
            <v-col cols="12" v-show="!showMap" class="">
              <ThingList
                :limit="searchLimit"
                :offset="searchOffset"
                :type-thing="searchType"
                :created-by="searchCreatedBy"
                :search-keywords="searchKeywords"
                :inactivated="searchInactivated"
                :validated="searchValidated"
                @thing-error="thingGotErr"
                @thing-ok="thingGotSuccess"
              />
            </v-col>
          </v-row>
        </v-container>
      </template>
      <template v-else>
        <Login
          :msg="`Authentification ${APP_TITLE}:`"
          :backend="APP_TITLE"
          :disabled="!isNetworkOk"
          @login-ok="loginSuccess"
          @login-error="loginFailure"
        />
      </template>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive } from "vue"
import { useDisplay } from "vuetify"
import { isNullOrUndefined } from "@/tools/utils"
import { APP, APP_TITLE, DEV, HOME, getLog, BUILD_DATE, VERSION, BACKEND_URL, defaultAxiosTimeout } from "@/config"
import Login from "@/components/Login.vue"
import ThingList from "@/components/ThingList.vue"
import MapLausanne from "@/components/MapLausanne.vue"
import { Configuration, DefaultApi, TypeThingList } from "@/openapi-generator-cli_thing_typescript-axios"
import {
  getUserIsAdmin,
  getTokenStatus,
  clearSessionStorage,
  doesCurrentSessionExist,
  getLocalJwtTokenAuth,
  getUserId,
  getSessionId,
} from "@/components/Login"
const log = getLog(APP, 4, 2)
let myApi: DefaultApi
type LevelAlert = "error" | "success" | "warning" | "info" | undefined
const feedbackTimeError = 6000
const feedbackTimeWarning = 4000
const displaySize = reactive(useDisplay())
const showSearchCriteria = ref(true)
const showSettings = ref(false)
const showMap = ref(true)
const searchType = ref(0)
const arrListTypeThing: TypeThingList[] = reactive([])
const searchCreatedBy = ref(undefined)
const searchKeywords = ref(undefined)
const searchInactivated = ref(false)
const searchValidated = ref(undefined)
const searchLimit = ref(25)
const searchOffset = ref(0)
const rules = {
  required: (value) => !!value || "Obligatoire.",
  max20Chars: (value) => value.length <= 20 || "Max 20 caractères",
  minNumber1: (value) => value > 1 || "Minimum autorisé = 1",
  email: (value) => {
    const pattern =
      /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
    return pattern.test(value) || "Format e-mail invalide."
  },
}

const isUserAuthenticated = ref(false)
const isUserAdmin = ref(false)
const isNetworkOk = ref(true)
const drawer = ref(false)
const feedbackTimeout = ref(2000) // default display time 5sec
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

const getMapIcon = () => (showMap.value ? "mdi-earth-off" : "mdi-earth")
const getMapTitle = () => (showMap.value ? "cacher la carte" : "afficher la carte")
const getFilterIcon = () => (showSearchCriteria.value ? "mdi-filter-off" : "mdi-filter")
const getFilterTitle = () =>
  showSearchCriteria.value ? "cacher les critères de filtrage" : "afficher les critères de filtrage"

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
          displayFeedBack(`Problème réseau :${val.msg}`, "error", feedbackTimeError)
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
            displayFeedBack("Votre session a expiré !", "warning", feedbackTimeWarning)
            logout()
          }
          displayFeedBack(`Un problème est survenu avec votre session erreur: ${val.err}`, "error", feedbackTimeError)
        }
      })
      .catch((err) => {
        log.e("# getJwtToken() in catch ERROR err: ", err)
        displayFeedBack(`Il semble qu'il y a eu un problème réseau ! erreur: ${err}`, "error", feedbackTimeError)
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

const thingGotErr = (v: string) => {
  log.w(`# entering... val:${v} `)
  displayFeedBack(v, "error", feedbackTimeError)
}

const thingGotSuccess = (v: string) => {
  log.t(`# entering... val:${v} `)
  displayFeedBack(v, "success")
}

const mapClickHandler = (pos: number[]) => {
  log.t(`## entering... pos:${pos[0]}, ${pos[1]}`)
}
const clearFilters = () => {
  searchCreatedBy.value = undefined
  searchKeywords.value = undefined
  searchType.value = 0
  searchValidated.value = undefined
  searchInactivated.value = false
  searchOffset.value = 0
  searchLimit.value = 250
}

const initialize = () => {
  log.t(`# entering...  `)
  const token = getLocalJwtTokenAuth()
  const myConf = new Configuration({
    accessToken: token,
    baseOptions: { timeout: defaultAxiosTimeout, headers: { "X-Goeland-Token": getSessionId() } },
    basePath: BACKEND_URL + "/goapi/v1",
  })
  myApi = new DefaultApi(myConf)
  searchCreatedBy.value = getUserId()
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
  log.l("displaySize ", displaySize)

  window.addEventListener("online", () => {
    log.w("ONLINE AGAIN :)")
    isNetworkOk.value = true
    displayFeedBack('⚡⚡🚀  LA CONNEXION RESEAU EST RÉTABLIE :  😊 vous êtes "ONLINE"  ', "success")
  })
  window.addEventListener("offline", () => {
    log.w("OFFLINE :((")
    isNetworkOk.value = false
    displayFeedBack('⚡⚡⚠ PAS DE RESEAU ! ☹ vous êtes "OFFLINE" ', "error", feedbackTimeError)
  })
})
</script>
