/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import "@mdi/font/css/materialdesignicons.css"
import "vuetify/styles"
import { md3 } from "vuetify/blueprints"
// Composables
import { createVuetify } from "vuetify"
//import { VDataTable } from "vuetify/labs/VDataTable"
//import { VDatePicker } from "vuetify/labs/VDatePicker"
//import { VuetifyDateAdapter } from "vuetify/labs/date/adapters/vuetify"
// Translations provided by Vuetify
import { fr, en } from "vuetify/locale"

const myCustomLightTheme = {
  dark: false,
  colors: {
    background: "#FFFFFF",
    surface: "#FFFFFF",
    "primary-lighten-1": "#7986CB",
    primary: "#3F51B5",
    "primary-darken-1": "#303F9F",
    "primary-accent-1": "#304FFE",
    "secondary-lighten-1": "#64B5F6",
    secondary: "#2196F3",
    "secondary-darken-1": "#1976D2",
    "secondary-accent-1": "#2962FF",
    error: "#D50000",
    info: "#03A9F4",
    success: "#4CAF50",
    warning: "#FF9800",
  },
}

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  /* theme: {
    defaultTheme: "myCustomLightTheme",
    themes: {
      myCustomLightTheme,
    },
  },
  blueprint: md3,
  */
  components: {},
  date: {
    //adapter: VuetifyDateAdapter,
  },
  locale: {
    locale: "fr",
    fallback: "en",
    messages: { fr, en },
  },
})
