<template>
  <v-app>
    <v-app-bar color="primary" density="compact">
      <v-app-bar-nav-icon variant="text" @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ `${APP_TITLE} v${VERSION}` }}</v-toolbar-title>
      <template v-if="DEV">
        <span class="left-0">{{ displaySize.name }}. TypeThings[{{ numTypeThings }}] areWeReady:{{ areWeReady }}</span>
      </template>
      <v-spacer></v-spacer>
      <template v-if="isUserAuthenticated">
        <v-btn variant="text" :icon="getMapIcon()" :title="getMapTitle()" @click="showMap = !showMap"></v-btn>
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
        v-model="appStore.feedbackVisible"
        :timeout="appStore.feedbackTimeout"
        rounded="pill"
        :color="appStore.feedbackType"
        location="top"
      >
        <v-alert
          class="ma-4"
          :type="appStore.feedbackType"
          :text="appStore.feedbackMsg"
          :color="appStore.feedbackType"
        ></v-alert>
      </v-snackbar>
      <template v-if="isUserAuthenticated">
        <div class="text-center">
          <v-overlay v-model="busyDoingNetWork" class="align-center justify-center">
            <v-progress-circular color="primary" indeterminate size="64"></v-progress-circular>
            <v-alert>
              <v-alert-title>
                Chargement des données... isInitDone: {{ store.isInitDone }},areWeReady:{{ areWeReady }}
              </v-alert-title>
            </v-alert>
          </v-overlay>
        </div>
        <v-card v-show="showSearchCriteria" variant="elevated" elevation="14" class="mx-4">
          <v-card-item>
            <v-toolbar color="transparent" class="px-0">
              <v-icon>mdi-magnify</v-icon>
              <v-toolbar-title>Critères de filtrage...</v-toolbar-title>
              <v-spacer></v-spacer>
              <v-btn dark color="primary" variant="flat" prepend-icon="mdi-eraser" @click.prevent="clearFilters"
                >Réinitialiser Filtres</v-btn
              >
              <template #extension>
                <v-tabs v-model="tabs" color="primary" grow>
                  <v-tab :value="1">
                    <v-icon>mdi-filter</v-icon>
                  </v-tab>

                  <v-tab :value="2">
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-tab>
                </v-tabs>
              </template>
            </v-toolbar>

            <v-window v-model="tabs">
              <v-window-item :value="1">
                <v-card>
                  <v-card-text>
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
                          :items="store.arrListTypeThing"
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
                  </v-card-text>
                </v-card>
              </v-window-item>
              <v-window-item :value="2">
                <v-card>
                  <v-card-text>
                    <v-row>
                      <v-col cols="12" sm="6" md="4" lg="4" xl="4">
                        <v-text-field
                          type="number"
                          :rules="[rules.required, rules.minNumber1, rules.maxNumber1]"
                          min="1"
                          max="maxSearchLimit"
                          v-model="searchLimit"
                          density="compact"
                          label="Max rows"
                          hint="Le nombre maximum d'enregistrements à récupérer dans la Base de données"
                        />
                      </v-col>
                      <v-col cols="12" sm="6" md="4" lg="4" xl="4">
                        <v-text-field
                          type="number"
                          v-model="searchOffset"
                          density="compact"
                          min="0"
                          label="Offset row"
                        />
                      </v-col>
                    </v-row>
                  </v-card-text>
                </v-card>
              </v-window-item>
            </v-window>
          </v-card-item>
        </v-card>

        <v-container class="text-center fill-height pa-2" :fluid="true">
          <v-row class="fill-height h-90">
            <v-col cols="12" class="">
              <ThingList
                :limit="+searchLimit"
                :offset="+searchOffset"
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
          <v-row class="fill-height h-100">
            <v-col cols="12" v-show="showMap" class="fill-height">
              <MapLausanne :zoom="3" @map-click="mapClickHandler"></MapLausanne>
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
import { APP, APP_TITLE, DEV, HOME, getLog, BUILD_DATE, VERSION } from "@/config"
import { useAppStore } from "@/appStore"
import Login from "@/components/Login.vue"
import { useThingStore, ISearchThingParameters, maxSearchLimit, defaultQueryLimit } from "@/components/ThingStore"
import ThingList from "@/components/ThingList.vue"
import MapLausanne from "@/components/MapLausanne.vue"
import {
  getUserIsAdmin,
  getTokenStatus,
  clearSessionStorage,
  doesCurrentSessionExist,
  getUserId,
} from "@/components/Login"
import { mapClickInfo } from "@/components/Map"
import { storeToRefs } from "pinia"

const log = getLog(APP, 4, 2)
const appStore = useAppStore()
const defaultFeedbackErrorTimeout = 5000 // default display time 5sec
const displaySize = reactive(useDisplay())
const showSearchCriteria = ref(true)
const showSettings = ref(false)
const showMap = ref(false)
const searchType = ref(0)
const searchCreatedBy = ref(0)
const searchKeywords = ref(undefined)
const searchInactivated = ref(false)
const searchValidated = ref(undefined)
const searchLimit = ref(defaultQueryLimit)
const searchOffset = ref(0)
const store = useThingStore()
const { areWeReady, busyDoingNetWork, numTypeThings } = storeToRefs(store)
const tabs = ref(null)
const rules = {
  required: (value) => !!value || "Obligatoire.",
  max20Chars: (value) => value.length <= 20 || "Max 20 caractères",
  minNumber1: (value) => value > 1 || "Minimum autorisé = 1",
  maxNumber1: (value) => value <= maxSearchLimit || `Maximum autorisé = ${maxSearchLimit}`,
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
let autoLogoutTimer: number

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
  appStore.displayFeedBack("Vous vous êtes déconnecté de l'application avec succès !", "success")
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
          appStore.displayFeedBack(`Problème réseau :${val.msg}`, "error", defaultFeedbackErrorTimeout)
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
            appStore.displayFeedBack("Votre session a expiré !", "warning", defaultFeedbackErrorTimeout)
            logout()
          }
          appStore.displayFeedBack(
            `Un problème est survenu avec votre session erreur: ${val.err}`,
            "error",
            defaultFeedbackErrorTimeout
          )
        }
      })
      .catch((err) => {
        log.e("# getJwtToken() in catch ERROR err: ", err)
        appStore.displayFeedBack(
          `Il semble qu'il y a eu un problème réseau ! erreur: ${err}`,
          "error",
          defaultFeedbackErrorTimeout
        )
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
  appStore.hideFeedBack()
  appStore.displayFeedBack("Vous êtes authentifié sur l'application.", "success")
  if (isNullOrUndefined(autoLogoutTimer)) {
    // check every 600 seconds(600'000 milliseconds) if jwt is still valid
    autoLogoutTimer = window.setInterval(checkIsSessionTokenValid, 600000)
  }
  initialize()
}

const loginFailure = (v: string) => {
  log.w(`# entering... val:${v} `)
  isUserAuthenticated.value = false
  isUserAdmin.value = false
}

const thingGotErr = (v: string) => {
  log.w(`# entering... val:${v} `)
  appStore.displayFeedBack(v, "error", defaultFeedbackErrorTimeout)
}

const thingGotSuccess = (v: string) => {
  log.t(`# entering... val:${v} `)
  appStore.displayFeedBack(v, "success")
}

const mapClickHandler = (clickInfo: mapClickInfo) => {
  log.t(`## entering... pos:${clickInfo.x}, ${clickInfo.y}`)
  log.t(`##features length :${clickInfo.features.length}`, clickInfo.features)
}
const clearFilters = () => {
  log.t(`# App.vue clearFilters  `)
  searchCreatedBy.value = 0
  searchKeywords.value = undefined
  searchType.value = 0
  searchValidated.value = undefined
  searchInactivated.value = false
  searchOffset.value = 0
  searchLimit.value = defaultQueryLimit
}

const initialize = async () => {
  log.t(`# App.vue entering initialize...  `)
  searchCreatedBy.value = getUserId()
  const initialSearchParameters = Object.assign({}, {
    createdBy: searchCreatedBy.value,
    searchKeywords: searchKeywords.value,
    typeThing: searchType.value,
    validated: searchValidated.value,
    inactivated: searchInactivated.value,
    limit: searchLimit.value,
    offset: searchOffset.value,
  } as ISearchThingParameters)
  if (!store.isInitDone) {
    await store.init(initialSearchParameters)
    log.l(`## Initialize in ThingListVue Done, arrListTypeThing length : ${numTypeThings}`)
  }
}

onMounted(() => {
  log.l(`Main App.vue ${APP}-${VERSION}, du ${BUILD_DATE}`)

  window.addEventListener("online", () => {
    log.w("ONLINE AGAIN :)")
    appStore.networkOnLine()
    appStore.displayFeedBack('⚡⚡🚀  LA CONNEXION RESEAU EST RÉTABLIE :  😊 vous êtes "ONLINE"  ', "success")
  })
  window.addEventListener("offline", () => {
    log.w("OFFLINE :((")
    appStore.networkOffLine()
    appStore.displayFeedBack('⚡⚡⚠ PAS DE RESEAU ! ☹ vous êtes "OFFLINE" ', "error", defaultFeedbackErrorTimeout)
  })
})
</script>
