<style lang="scss">
@import "ol/ol.css";
@import "ol-layerswitcher/dist/ol-layerswitcher.css";
$searchbox_height: 4.85em; // = 64px (body font size =?)
$searchbox_height_16px: 4em; // = 64px (body font size =16px)
$button_size_14px: 3em; // = 42px (body font size = 14px)
$button_size_20px: 2.1em; // = 42px (body font size = 20px)
.map {
  background-color: white;
  position: absolute;
  top: 0;
  bottom: 0;
  width: 100%;
  height: 99%;
  min-height: 450px;
}
.mouse-coordinates {
  position: absolute;
  top: 10px;
  left: 50px;
  z-index: 250;
}

.searchBox {
  // .v-input font-size: 16px; (= body font size)
  padding-left: 17px;
  padding-right: 17px;
  height: $searchbox_height_16px; // = 64px
  top: 15px;
}

.ol-control:hover {
  background-color: rgba(0, 0, 0, 0);
}

.ol-control {
  font-size: 18px;

  button {
    background-color: rgba(245, 245, 245, 1);
    color: black;
    font-weight: normal;
    box-shadow:
      0px 3px 1px -2px rgba(0, 0, 0, 0.2),
      0px 2px 2px 0px rgba(0, 0, 0, 0.14),
      0px 1px 5px 0px rgba(0, 0, 0, 0.12);
    transition-property:
      box-shadow,
      transform,
      opacity,
      -webkit-box-shadow,
      -webkit-transform;
    border-radius: 4px;
  }

  button:hover {
    background-color: rgba(245, 245, 245, 1);
    color: black;
  }

  button:focus {
    background-color: rgba(245, 245, 245, 1);
    color: black;
  }
}
.ol-zoom {
  top: calc($button_size_20px/2); // = 1.05em = 21px
  left: unset !important;
  right: 0.5em;
  background-color: rgba(255, 255, 255, 0);

  .ol-zoom-in {
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }

  .ol-zoom-out {
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }
}
.layers_button {
  // .v-btn.v-size--default font-size: 0.875rem = 14px (= body font size)
  margin-right: -0.3em; // -4px
  top: calc($searchbox_height + (3 * $button_size_20px) + 0.2em); // 197px
  left: unset !important;
  right: 0.5em;
  z-index: 250;
}

.layer-switcher-dialog {
  // min-height: 350px;
  max-width: 250px;
  padding: 10px;
  ul {
    list-style: none;
  }

  li {
    padding-top: 0.5em;
    padding-left: 0.1em;
    text-indent: -1.5em;
  }

  label {
    padding-left: 10px;
    vertical-align: bottom;
  }
}
.gps_button {
  // .v-btn.v-size--default font-size: 0.875rem = 14px (= body font size)
  margin-right: -0.3em; // -4px
  top: $searchbox_height + (4.5*$button_size_14px) + 0.2; // 260px
}

.ol-attribution {
  bottom: 1em;
  margin-right: 0.15em; // 3px
  font-size: 0.8em;
  position: fixed;
  background-color: rgba(255, 255, 255, 0);
}
</style>
<template>
  <v-responsive class="d-flex fill-height">
    <!-- TEST RENDU LAYERSWITCHER PERSONALISE -->
    <div class="text-xs-center">
      <v-dialog v-model="layerSwitcherVisible" eager width="290">
        <v-card>
          <v-card-title class="subtitle-1 grey lighten-2 pl-6 pt-2 pb-1" primary-title>
            Choix des couches
          </v-card-title>

          <v-card-text>
            <div id="divLayerSwitcher" class="layer-switcher-dialog"></div>
          </v-card-text>

          <v-divider></v-divider>

          <v-card-actions class="pa-1">
            <v-spacer></v-spacer>
            <v-btn
              class="caption white--text"
              color="indigo lighten-1"
              height="25"
              @click="layerSwitcherVisible = false"
            >
              Fermer
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>
    <v-btn
      outlined
      icon="mdi-layers-outline"
      aria-label="selection couches"
      class="layers_button"
      height="42"
      min-width="42"
      right
      width="42"
      @click.stop="toggleLayerSwitcher"
    ></v-btn>
    <v-sheet class="mouse-coordinates" :elevation="10" :height="50" :width="400" border rounded>
      <div>x,y: {{ posMouseX }}, {{ posMouseY }}</div>
      <div>Propriétés:{{ propsValues }}</div>
    </v-sheet>
    <div class="map" id="map" ref="myMap">
      <noscript> You need to have a browser with javascript support to see this OpenLayers map of Lausanne </noscript>
    </div>
  </v-responsive>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, watch } from "vue"
import { getLog } from "@/config"
import { createLausanneMap } from "@/components/Map"
import OlMap from "ol/Map"
import LayerSwitcher from "ol-layerswitcher"

const log = getLog("ThingListVue", 4, 2)
const posMouseX = ref(0)
const posMouseY = ref(0)
const layerSwitcherVisible = ref(false)
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
const toggleLayerSwitcher = () => {
  log.t(`# toggleLayerSwitcher layerSwitcherVisible : ${layerSwitcherVisible.value}`)
  if (layerSwitcherVisible.value) {
    layerSwitcherVisible.value = false
  } else {
    layerSwitcherVisible.value = true
  }
}
const initialize = async (center) => {
  log.t(" #> entering initialize ...")
  myOlMap = await createLausanneMap("map", center, 8, "fonds_geo_osm_bdcad_couleur")
  if (myOlMap !== null) {
    myOlMap.on("pointermove", (evt) => {
      posMouseX.value = +Number(evt.coordinate[0]).toFixed(2)
      posMouseY.value = +Number(evt.coordinate[1]).toFixed(2)
    })
    const divToc = document.getElementById("divLayerSwitcher")
    LayerSwitcher.renderPanel(myOlMap, divToc, {})
  }
}

onMounted(() => {
  log.t("mounted()")
  const placeStFrancoisM95 = [2538202, 1152364]
  initialize(placeStFrancoisM95)
})
</script>
