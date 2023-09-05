<template>
  <v-container class="fill-height">
    <v-responsive class="d-flex align-center text-center fill-height">
      <v-row class="d-flex align-center justify-center">
        <v-col cols="auto">
          <v-card class="elevation-12">
            <v-toolbar dark color="primary">
              <v-toolbar-title>{{ title }}</v-toolbar-title>
              <v-spacer></v-spacer>
            </v-toolbar>
            <v-card-text>
              <v-form ref="login-form" v-model="validLoginForm" lazy-validation>
                <v-text-field
                  id="username"
                  name="username"
                  ref="username"
                  v-model="username"
                  required
                  autocomplete="username"
                  prepend-icon="person"
                  :cleareable="true"
                  :rules="[() => !!username || 'Veuillez saisir votre login, il est obligatoire']"
                  @keyup.enter="onEnterKey"
                  label="Entrez votre utilisateur"
                  type="text"
                >
                </v-text-field>
                <!-- add this to get chrome autocomplete
                browser-autocomplete="current-password"-->
                <v-text-field
                  id="password"
                  name="password"
                  ref="password"
                  v-model="password"
                  required
                  autocomplete="current-password"
                  prepend-icon="lock"
                  :rules="[() => !!password || 'Veuillez saisir votre mot de passe, il est obligatoire']"
                  label="votre mot de passe"
                  :append-icon="showPassword ? 'visibility_off' : 'visibility'"
                  @click:append="showPassword = !showPassword"
                  @keyup.enter="onEnterKey"
                  :type="showPassword ? 'text' : 'password'"
                >
                </v-text-field>
                <v-text-field id="password_hash" name="password_hash" ref="password_hash" v-show="sha256Visible" v-model="password_hash" prepend-icon="lock" label="sha256" type="text"> </v-text-field>
              </v-form>
            </v-card-text>
            <v-card-actions>
              <v-alert :value="feedbackVisible" :color="feedbackType" :icon="feedbackType" outlined>
                {{ feedbackText }}
              </v-alert>
              <v-spacer></v-spacer>
              <v-btn color="primary" @click.prevent="getJwtToken">Connexion</v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script>
import { isNullOrUndefined } from "@/tools/utils"
import { APP, BACKEND_URL, getLog } from "../config"
import { getPasswordHash, getToken } from "./Login"

const log = getLog("Login-Vue", 2, 2)
export default {
  name: "LoginVue",
  data: () => ({
    drawer: null,
    username: "bill",
    password: null,
    showPassword: false,
    sha256Visible: false,
    validLoginForm: false,
    feedbackVisible: false,
    feedbackText: "Veuillez vous authentifier SVP.",
    feedbackType: "info",
  }),

  props: {
    base_server_url: {
      type: String,
      default: BACKEND_URL,
    },
    title: {
      type: String,
      default: `Authentification ${APP}`,
    },
  },

  computed: {
    password_hash() {
      return getPasswordHash(this.password)
    },
  },
  mounted() {
    log.t("# IN mounted()")
    this.$refs.username.focus()
  }, // end of mounted
  methods: {
    onEnterKey() {
      log.t("# IN onEnterKey()")
      if (this.username.trim().length < 1) {
        this.displayFeedBack("Veuillez saisir votre utilisateur, il est obligatoire!")
        this.$refs.username.focus()
        return false
      }
      if (this.password.trim().length < 1) {
        this.displayFeedBack("Veuillez saisir votre mot de passe, il est obligatoire!")
        return false
      }
      this.getJwtToken()
      return true
    },
    displayFeedBack(text, type) {
      this.feedbackText = text
      this.feedbackType = type
      this.feedbackVisible = true
    },
    resetFeedBack() {
      this.feedbackText = ""
      this.feedbackType = "info"
      this.feedbackVisible = false
    },
    getJwtToken() {
      log.t("# IN getJwtToken()")
      this.resetFeedBack()
      if (this.$refs["login-form"].validate()) {
        try {
          const res = getToken(this.base_server_url, this.username, this.password_hash)
            .then((val) => {
              if (val.data == null) {
                if (!isNullOrUndefined(val.err)) {
                  log.e(`# getJwtToken() ${val.msg}, ERROR is: `, val.err)
                  this.displayFeedBack(`Problème réseau :${val.msg}`, "error")
                }
                log.w(`# getToken received status ${val.status}`, val)
                this.displayFeedBack("Vos informations de connexions sont erronées !", "warning")
                this.$emit("loginError", "LOGIN FAILED", val.err)
              } else {
                log.l("# getJwtToken() SUCCESS val.data: ", val.data)
                this.$emit("login-ok", "LOGIN SUCCESS")
              }
            })
            .catch((err) => {
              log.e("# getJwtToken() in catch ERROR err: ", err)
              this.displayFeedBack(`Il semble qu'il y a un problème de réseau !${err}`, "error")
              this.$emit("loginError", "LOGIN ERROR", err)
            })
          log.l("# getJwtToken() after getToken res:", res)
        } catch (e) {
          log.t("# getJwtToken() TRY CATCH ERROR : ", e)
        }
      } else if (this.username.trim().length < 1) {
        this.displayFeedBack("Veuillez saisir votre utilisateur, il est obligatoire!")
      } else if (this.password.trim().length < 1) {
        this.displayFeedBack("Veuillez saisir votre mot de passe, il est obligatoire!")
      }
      log.t("# GOING OUT getJwtToken()")
    },
  },
}
</script>
