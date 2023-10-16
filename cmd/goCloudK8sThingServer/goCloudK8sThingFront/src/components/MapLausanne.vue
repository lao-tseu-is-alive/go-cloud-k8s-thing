<style lang="scss">
@import "ol/ol.css";
@import "ol-layerswitcher/dist/ol-layerswitcher.css";
.map {
  background-color: white;
  position: absolute;
  top: 0;
  bottom: 0;
  width: 100%;
  height: 99%;
  min-height: 450px;
}
</style>
<template>
  <v-responsive class="d-flex fill-height">
    <v-row>
      <v-col cols="6">x,y: {{ posMouseX }}, {{ posMouseY }}</v-col>
      <v-col cols="6">Propriétés:{{ propsValues }}</v-col>
    </v-row>
    <v-row class="d-flex fill-height">
      <v-col cols="12">
        <div class="map" id="map" ref="myMap">
          <noscript>
            You need to have a browser with javascript support to see this OpenLayers map of Lausanne
          </noscript>
        </div>
      </v-col>
    </v-row>
  </v-responsive>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, watch } from "vue"
import { getLog } from "@/config"
import { createLausanneMap } from "@/components/Map"
import OlMap from "ol/Map"

const log = getLog("ThingListVue", 4, 2)
const posMouseX = ref(0)
const posMouseY = ref(0)
const areWeReady = ref(false)
let myOlMap: null | OlMap
const myMap = ref(null)
const myProps = defineProps<{
  zoom?: number | undefined
}>()

//// EVENT SECTION

//const emit = defineEmits(["thing-ok", "thing-error"])

//// WATCH SECTION
watch(
  () => myProps.zoom,
  (val, oldValue) => {
    log.t(` watch myProps.zoom old: ${oldValue}, new val: ${val}`)
    if (val !== undefined && areWeReady.value == true) {
      if (val !== oldValue) {
        // do something
      }
    }
  }
  //  { immediate: true }
)
//// COMPUTED SECTION

const propsValues = computed(() => {
  return JSON.stringify(myProps, undefined, 3)
})

//// FUNCTIONS SECTION

const initialize = async (center) => {
  log.t(" #> entering initialize ...")
  myOlMap = await createLausanneMap("map", center, 8, "fonds_geo_osm_bdcad_couleur")
  myOlMap.on("pointermove", (evt) => {
    posMouseX.value = +Number(evt.coordinate[0]).toFixed(2)
    posMouseY.value = +Number(evt.coordinate[1]).toFixed(2)
  })
}

onMounted(() => {
  log.t("mounted()")
  const placeStFrancoisM95 = [2538202, 1152364]
  initialize(placeStFrancoisM95)
})
</script>
