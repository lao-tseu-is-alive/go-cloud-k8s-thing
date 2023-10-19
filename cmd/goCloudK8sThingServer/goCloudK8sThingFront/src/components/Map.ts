/**
 * Created by cgil on 2022-03-30.
 *
 * v.2.0.0 : Migration to TypeScript on 2021-12-21.
 */
import proj4 from "proj4"
import OlMap from "ol/Map"
import OlView from "ol/View"
import OlProjection from "ol/proj/Projection"
import OlLayerTile from "ol/layer/Tile"
import OlFormatWMTSCapabilities from "ol/format/WMTSCapabilities"
import OlSourceWMTS, { optionsFromCapabilities, Options } from "ol/source/WMTS"
import { register } from "ol/proj/proj4"
import LayerSwitcher from "ol-layerswitcher"
import { getLog } from "@/config"

const log = getLog("Login", 4, 2)

const urlLausanneMN95 = "https://tilesmn95.lausanne.ch/tiles/1.0.0/LausanneWMTS.xml"
const MaxExtent = [2532500, 1149000, 2545625, 1161000]
const lausanneGare = [2537968.5, 1152088.0]

proj4.defs(
  "EPSG:2056",
  "+proj=somerc +lat_0=46.95240555555556 +lon_0=7.439583333333333 +k_0=1 +x_0=2600000 +y_0=1200000 +ellps=bessel +towgs84=674.374,15.056,405.346,0,0,0,0 +units=m +no_defs"
)
proj4.defs(
  "EPSG:21781",
  "+proj=somerc +lat_0=46.95240555555556 +lon_0=7.439583333333333 +k_0=1 +x_0=600000 +y_0=200000 +ellps=bessel +towgs84=674.4,15.1,405.3,0,0,0,0 +units=m +no_defs"
)

register(proj4)
const swissProjection = new OlProjection({
  code: "EPSG:2056",
  extent: MaxExtent,
  units: "m",
})
const parser = new OlFormatWMTSCapabilities()

async function getWMTSCapabilitiesFromUrl(url: string) {
  const response = await fetch(url)
  if (!response.ok) {
    const message = `###!### ERROR in getWMTSCapabilitiesFromUrl when doing fetch(${url}: http status: ${response.status}`
    throw new Error(message)
  }
  const WMTSCapabilities = await response.text()
  return WMTSCapabilities
}

function getWmtsSource(WMTSCapabilitiesParsed, layerName: string) {
  const localDebug = false
  if (localDebug) log.t(`layerName: ${layerName}`)
  const WMTSOptions = optionsFromCapabilities(WMTSCapabilitiesParsed, {
    layer: layerName,
    matrixSet: "EPSG2056",
    format: "image/png",
    style: "default",
    crossOrigin: "anonymous",
  })
  return new OlSourceWMTS(<Options>WMTSOptions)
}

function createBaseOlLayerTile(parsedWmtsCapabilities, title: string, layerName: string, visible = false) {
  return new OlLayerTile({
    title: title,
    type: "base",
    visible,
    source: getWmtsSource(parsedWmtsCapabilities, layerName),
  })
}

async function getWmtsBaseLayers(url: string, defaultBaseLayer: string) {
  const arrWmtsLayers = []
  try {
    const WMTSCapabilities = await getWMTSCapabilitiesFromUrl(url)

    const WMTSCapabilitiesParsed = parser.read(WMTSCapabilities)
    console.log(`## in getWmtsBaseLayers(${url} : WMTSCapabilitiesParsed : \n`, WMTSCapabilitiesParsed)
    arrWmtsLayers.push(
      createBaseOlLayerTile(
        WMTSCapabilitiesParsed,
        "Orthophoto 2016 (Lausanne)",
        "orthophotos_ortho_lidar_2016",
        defaultBaseLayer === "orthophotos_ortho_lidar_2016"
      )
    )
    arrWmtsLayers.push(
      createBaseOlLayerTile(
        WMTSCapabilitiesParsed,
        "Fond cadastral (Lausanne)",
        "fonds_geo_osm_bdcad_gris",
        defaultBaseLayer === "fonds_geo_osm_bdcad_gris"
      )
    )
    arrWmtsLayers.push(
      createBaseOlLayerTile(
        WMTSCapabilitiesParsed,
        "Plan ville (Lausanne)",
        "fonds_geo_osm_bdcad_couleur",
        defaultBaseLayer === "fonds_geo_osm_bdcad_couleur"
      )
    )
    return arrWmtsLayers
  } catch (err) {
    const message = `###!### ERROR in getWmtsBaseLayers occured with url:${url}: error is: ${err}`
    console.warn(message)
    return []
  }
}
/**
 * check if the given f is a Function
 * @param divOfMap the id of the div you want to draw a map
 * @param centerOfMap the position where you want to center map in MN95 Coordinates [x,y] array
 * @param zoomLevel
 * @param baseLayer one of orthophotos_ortho_lidar_2016 fonds_geo_osm_bdcad_(gris|couleur)
 * @param divLayerSwitcherContent is the reference to the div where you want ot render the layer list
 * @returns an instance of an OpenLayer Map
 */
export async function createLausanneMap(
  divOfMap: string,
  centerOfMap = lausanneGare,
  zoomLevel = 16,
  baseLayer = "fonds_geo_osm_bdcad_couleur"
) {
  log.t(`createLausanneMap(x,y: [${centerOfMap[0]},${centerOfMap[1]}]  zoom:${zoomLevel})`)
  const arrBaseLayers = await getWmtsBaseLayers(urlLausanneMN95, baseLayer)
  if (arrBaseLayers === null || arrBaseLayers.length < 1) {
    log.w("arrBaseLayers cannot be null or empty to be able to see a nice map !")
    return null
  }
  const map = new OlMap({
    target: divOfMap,
    layers: arrBaseLayers,
    view: new OlView({
      projection: swissProjection,
      center: centerOfMap,
      zoom: zoomLevel,
    }),
  })
  const layerSwitcher = new LayerSwitcher({
    activationMode: "click",
    tipLabel: "Afficher la liste des fonds de plan", // Optional label for button
    collapseTipLabel: "Cacher la liste des fonds de plan", // Optional label for button
    groupSelectStyle: "children", // Can be 'children' [default], 'group' or 'none'
  })
  // map.addControl(layerSwitcher)
  // MAP EVENTS
  map.on("click", async (evt) => {
    const x = Number(evt.coordinate[0])
    const y = Number(evt.coordinate[1])
    log.l(`#in olmap.click at x,y : [${x.toFixed(2)}, ${y.toFixed(2)}]\n`)
  })
  return map
}
