/**
 * plugins/vuetify.ts
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import "@mdi/font/css/materialdesignicons.css"
import "vuetify/styles"
import { md2 } from "vuetify/blueprints"
// Composables
import { createVuetify } from "vuetify"
import { VDataTable } from "vuetify/labs/VDataTable"

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  theme: {
    themes: {
      light: {
        colors: {
          primary: "#1867C0",
          secondary: "#5CBBF6",
        },
      },
    },
  },
  blueprint: md2,
  components: {
    VDataTable,
  },
})
