<style lang="scss">
@import "ol/ol.css";
@import "ol-layerswitcher/dist/ol-layerswitcher.css";
$searchbox_height: 4.85em; // = 64px (body font size =?)
$searchbox_height_16px: 4em; // = 64px (body font size =16px)
$button_size_14px: 3em; // = 42px (body font size = 14px)
$button_size_20px: 2.1em; // = 42px (body font size = 20px)
.map {
  background-color: white;
  position: relative;
  top: 0px;
  bottom: 0px;
  width: 100%;
  height: 98%;
  min-height: 450px;
}

.tooltip {
  position: relative;
  padding: 3px;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  opacity: 0.8;
  white-space: nowrap;
  font: 12pt sans-serif;
}
.mouse-coordinates {
  z-index: 245;
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
  position: absolute;
  top: calc($button_size_20px/2 + 2em); // = 1.05em = 21px
  left: unset !important;
  right: 0.5em;
  margin-bottom: 1em;
  background-color: rgba(255, 255, 255, 0);

  .ol-zoom-in {
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }

  .ol-zoom-out {
    top: calc($button_size_20px/2 + $button_size_20px + 1.5em); // = 1.05em = 21px
    height: 42px;
    width: 42px;
    min-width: 42px;
    color: rgba(0, 0, 0, 0.87);
    border-radius: 4px;
  }
}
.layers_button {
  // .v-btn.v-size--default font-size: 0.875rem = 14px (= body font size)
  position: absolute;
  top: calc($searchbox_height + (3 * $button_size_20px) + 1.5em); // 197px
  //left: unset !important;
  right: 0.5em;
  z-index: 250;
  background-color: #feffeb;
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
  top: calc($searchbox_height + (4.5 * $button_size_14px) + 0.2em); // 260px
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
      icon="mdi-layers-outline"
      aria-label="selection couches"
      color="white"
      class="layers_button"
      height="45"
      min-width="45"
      width="45"
      @click.stop="toggleLayerSwitcher"
    ></v-btn>
    <v-footer class="text-center bottom--1 mouse-coordinates">
      <div>x,y: {{ posMouseX }}, {{ posMouseY }}</div>
      <div>Trouvé {{ store.numRecords }} Thing(s) avec ces filtres:{{ searchParameters }} </div>
      &nbsp;
    </v-footer>
    <div class="map" id="map" ref="myMap">
      <noscript> You need to have a browser with javascript support to see this OpenLayers map of Lausanne </noscript>
      <div ref="mapTooltip" class="tooltip"></div>
    </div>
  </v-responsive>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue"
import { getLog } from "@/config"
import { addGeoJsonLayer, createLausanneMap, mapClickInfo, mapFeatureInfo } from "@/components/Map"
import OlMap from "ol/Map"
import OlOverlay from "ol/Overlay"
import LayerSwitcher from "ol-layerswitcher"
import { geoData } from "@/components/geodata"
import { isNullOrUndefined } from "@/tools/utils"
import { useThingStore } from "@/components/ThingStore"
import { storeToRefs } from "pinia"

const store = useThingStore()
const { searchParameters } = storeToRefs(store)

const log = getLog("ThingListVue", 4, 2)
const myLayerName = "GoelandThingLayer"
const posMouseX = ref(0)
const posMouseY = ref(0)
const layerSwitcherVisible = ref(false)
let myOlMap: null | OlMap
let myMapOverlay: null | OlOverlay
const mapTooltip = ref<HTMLDivElement | null>(null)
const myProps = defineProps<{
  zoom?: number | undefined
}>()

//// EVENT SECTION

const emit = defineEmits(["map-click", "map-error"])

//// WATCH SECTION
watch(
  () => myProps.zoom,
  (val, oldValue) => {
    log.t(` watch myProps.zoom old: ${oldValue}, new val: ${val}`)
    if (val !== undefined) {
      if (val !== oldValue) {
        // do something
      }
    }
  }
  //  { immediate: true }
)
//// COMPUTED SECTION

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
      const features = []
      if (myOlMap instanceof OlMap) {
        myOlMap.forEachFeatureAtPixel(evt.pixel, (feature, layer) => {
          let layerName = ""
          if (!isNullOrUndefined(layer)) {
            layerName = layer.get("name")
          }
          // on veut les tooltip seulement pour la couche myLayerName
          if (!isNullOrUndefined(layerName)) {
            if (layerName.indexOf(myLayerName) > -1) {
              const featureProps = feature.getProperties()
              if (!isNullOrUndefined(featureProps)) {
                const featureInfo = {
                  id: featureProps.id,
                  feature,
                  layer: layerName,
                  data: featureProps,
                }
                // log.l(`Feature id : ${feature_props.id}, info:`, info);
                features.push(featureInfo)
              } else {
                features.push({
                  id: 0,
                  feature,
                  layer: layerName,
                })
              }
            }
          }
          // return feature
        })
      } // end of forEachFeatureAtPixel
      if (features.length > 0) {
        //log.w("##MapLausanne pointermove EVENT ->Array of features found :", features, mapTooltip.value)
        let strToolTip = ""
        let layerName = ""
        features.forEach((featInfo) => {
          const currentTitle = featInfo.data.name
          layerName += `${featInfo.layer} `
          if (!isNullOrUndefined(currentTitle)) {
            strToolTip += `${currentTitle.replace(/(<([^>]+)>)/gi, "")}<br>`
          }
        })
        if (strToolTip.length > 0 && layerName.indexOf(myLayerName) > -1) {
          // log.w(`Before tooltip layer : ${layerName}`)
          if (myMapOverlay instanceof OlOverlay) {
            myMapOverlay.setPosition(evt.coordinate)
          }
          if (mapTooltip.value instanceof HTMLDivElement) {
            mapTooltip.value.style.display = "block"
            mapTooltip.value.innerHTML = strToolTip
          }
        } else {
          if (mapTooltip.value instanceof HTMLDivElement) {
            mapTooltip.value.style.display = "none"
            mapTooltip.value.innerHTML = ""
          }
        }
      } else {
        if (mapTooltip.value instanceof HTMLDivElement) {
          mapTooltip.value.style.display = "none"
          mapTooltip.value.innerHTML = ""
        }
      }
    })
    myOlMap.on("click", (evt) => {
      const x = +Number(evt.coordinate[0]).toFixed(2)
      const y = +Number(evt.coordinate[1]).toFixed(2)
      const features: mapFeatureInfo[] = []
      if (myOlMap instanceof OlMap) {
        myOlMap.forEachFeatureAtPixel(evt.pixel, (feature, layer) => {
          let layerName = ""
          if (!isNullOrUndefined(layer)) {
            layerName = layer.get("name")
          }
          // on veut les tooltip seulement pour la couche myLayerName
          if (!isNullOrUndefined(layerName)) {
            if (layerName.indexOf(myLayerName) > -1) {
              const featureProps = feature.getProperties()
              if (!isNullOrUndefined(featureProps)) {
                const featureInfo: mapFeatureInfo = {
                  id: featureProps.id,
                  // @ts-expect-error it's ok
                  feature,
                  layer: layerName,
                  data: featureProps,
                }
                // log.l(`Feature id : ${feature_props.id}, info:`, info);
                features.push(featureInfo)
              } else {
                // @ts-expect-error it's ok
                features.push({ id: 0, feature, layer: layerName, data: null } as mapFeatureInfo)
              }
            }
          }
          // return feature
        })
      } // end of forEachFeatureAtPixel
      if (features.length > 0) {
        features.forEach((featInfo) => {
          log.l("Feature found : ", featInfo)
        })
      }
      emit("map-click", { x, y, features } as mapClickInfo)
    })
    const divToc = document.getElementById("divLayerSwitcher")
    LayerSwitcher.renderPanel(myOlMap, divToc, {})
    addGeoJsonLayer(myOlMap, myLayerName, geoData)
    if (mapTooltip.value !== null) {
      myMapOverlay = new OlOverlay({
        element: mapTooltip.value as HTMLDivElement,
        offset: [0, -40],
        positioning: "top-center",
      })
      myOlMap.addOverlay(myMapOverlay)
    }
  }
}

onMounted(() => {
  log.t("mounted()")
  const placeStFrancoisM95 = [2538202, 1152364]
  initialize(placeStFrancoisM95)
})
</script>
