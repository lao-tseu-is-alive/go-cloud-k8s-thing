import { defineStore } from "pinia"
import { getLog, BACKEND_URL, defaultAxiosTimeout } from "@/config"
import { isNullOrUndefined, parseJsonWithDetailedError } from "@/tools/utils"
import { Thing, ThingList, TypeThingList } from "@/openapi-generator-cli_thing_typescript-axios"
import axios, { AxiosInstance, CreateAxiosDefaults } from "axios"
import { getLocalJwtTokenAuth, getSessionId } from "@/components/Login"

export const defaultQueryLimit = 100
const log = getLog("ThingStore", 4, 2)
let myAxios: AxiosInstance

export interface ISearchThingParameters {
  typeThing?: number | undefined
  searchKeywords?: string | undefined
  createdBy?: number | undefined
  inactivated: boolean
  validated?: boolean | undefined
  limit: number
  offset: number
}

export const maxSearchLimit: number = 1000
export const defaultListItem: ThingList = {
  id: crypto.randomUUID(),
  type_id: 0,
  name: "",
  description: undefined,
  external_id: undefined,
  inactivated: false,
  validated: undefined,
  status: undefined,
  created_by: 0,
  created_at: undefined,
  pos_x: 0,
  pos_y: 0,
}
export const defaultItem: Thing = {
  id: crypto.randomUUID(),
  type_id: 0,
  name: "",
  description: undefined,
  comment: undefined,
  external_id: undefined,
  external_ref: undefined,
  build_at: undefined,
  status: undefined,
  contained_by: undefined,
  contained_by_old: undefined,
  inactivated: false,
  inactivated_time: undefined,
  inactivated_by: undefined,
  inactivated_reason: undefined,
  validated: undefined,
  validated_time: undefined,
  validated_by: undefined,
  managed_by: undefined,
  created_at: undefined,
  created_by: 0,
  last_modified_at: undefined,
  last_modified_by: undefined,
  deleted: false,
  deleted_at: undefined,
  deleted_by: undefined,
  more_data: undefined,
  pos_x: 0,
  pos_y: 0,
}

type netThing = { data: Thing | null; err: Error | null }

interface IDictionary {
  [key: number]: string
}

export const useThingStore = defineStore("Thing", {
  state: () => {
    return {
      records: [] as ThingList[],
      arrListTypeThing: [] as TypeThingList[],
      dicoTypeThing: {} as IDictionary,
      dicoTypeThingIconPath: {} as IDictionary,
      searchParameters: { inactivated: false, limit: defaultQueryLimit, offset: 0 } as ISearchThingParameters,
      areWeReady: false,
      isThereAnError: false,
      lastErrorMessage: "",
      isInitDone: false,
    }
  },
  getters: {
    numRecords: (state) => state.records.length,
    numTypeThings: (state) => state.arrListTypeThing.length,
    busyDoingNetWork: (state) => !state.areWeReady,
    getGeoJson: (state) => {
      log.t(`> Entering getGeoJson.. records.length : ${state.records.length}`)
      // const startTime = performance.now()
      if (state.records.length > 0) {
        let myGeoJson = null
        let result = '{"type": "FeatureCollection", "features": ['
        state.records.forEach((r: ThingList) => {
          const feature = `
           {
            "type": "Feature",
            "geometry": {
              "type": "Point",
              "crs": {
                "type": "name",
                "properties": {
                  "name": "EPSG:2056"
                }
              },
              "coordinates": [${r.pos_x}, ${r.pos_y}]
              },
              "properties": {
                "id": "${r.id}",
                "type_id": ${r.type_id},
                "name": "${r.name}",
                "external_id": ${r.external_id},
                "icon_path": "${state.dicoTypeThingIconPath[r.type_id]}"
              }},`
          // log.l(feature)
          result += feature
        })
        if (result.endsWith(",")) {
          result = result.slice(0, -1)
        }
        result += "]}"
        try {
          myGeoJson = parseJsonWithDetailedError(result)
        } catch (e) {
          log.w(`> Error in getGeoJson.. JSON.parse(result) : ${e}`, result)
        }
        return myGeoJson
      }
      return { type: "FeatureCollection", features: [] }
    },
  },
  actions: {
    async get(id: string): Promise<netThing> {
      log.t(`> Entering getThing: ${id}`)
      this.clearError()
      this.areWeReady = false
      try {
        const resp = await myAxios.get("thing/" + id)
        log.l(`SUCCESS myAPi.get(id:${resp.data.id}`)
        if (resp.status == 200) {
          this.areWeReady = true
          return { data: resp.data, err: null }
        } else {
          this.areWeReady = true
          log.w("getThing got problem", resp)
          return { data: null, err: Error(`problem in getThing status : ${resp.status}, ${resp.statusText}`) }
        }
      } catch (error) {
        if (axios.isAxiosError(error)) {
          log.w(`Try Catch Axios ERROR message:${error.message}, error:`, error)
          if (error.response !== undefined && error.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(error.response.data.message) ? "" : error.response.data.message
            const msg = `getThing error : ${error.message}. Server says : ${srvMessage}`
            log.w(msg)
            return { data: null, err: Error(msg) }
          } else {
            return { data: null, err: Error(`getThing error : ${error.message}.`) }
          }
        } else {
          log.e("unexpected error: ", error)
          return { data: null, err: Error(`unexpected error: in getThing Try catch : ${error}`) }
        }
      }
    },
    async search(params: ISearchThingParameters) {
      log.t(`> Entering searchThing.. typeThing: ${params.typeThing}, createdBy: ${params.createdBy} `)
      this.clearError()
      const startTime = performance.now()
      this.areWeReady = false
      const clearRecords = (): void => {
        if (this.records.length > 0) {
          this.records.splice(0)
        }
      }
      const urlParams = getUrlParameters(params)
      clearRecords()
      const afterClearRecordsTime = performance.now()
      log.l(`>> in searchThing.. afterClearRecordsTime: ${Math.round(afterClearRecordsTime - startTime)} milliseconds `)
      try {
        const resp = await myAxios.get("thing/search" + urlParams)
        const afterAwaitMyAxiosGetTime = performance.now()
        log.l(
          `>> in searchThing.. afterAwaitMyAxiosGetTime: ${Math.round(afterAwaitMyAxiosGetTime - afterClearRecordsTime)} milliseconds `
        )
        log.l("myAxios.get(thing/search) : ")
        this.records = resp.data
        /* next loop takes 4560 milliseconds with 1000 rows (14 milliseconds with direct allocation
        resp.data.forEach((r: Thing) => {
          this.records.push(r)
        })
        */
        const afterRespDataForEachTime = performance.now()
        log.l(
          `>> in searchThing.. afterRespDataForEachTime: ${Math.round(afterRespDataForEachTime - afterAwaitMyAxiosGetTime)} milliseconds `
        )
        this.areWeReady = true
        return { data: resp.data, err: null }
      } catch (err) {
        this.areWeReady = true
        this.isThereAnError = true
        if (axios.isAxiosError(err)) {
          log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
          this.lastErrorMessage = err.message
          if (err.response !== undefined && err.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
            return { data: null, err: Error(`searchThing error : ${err.message}. Server says : ${srvMessage}`) }
          } else {
            return { data: null, err: Error(`searchThing error : ${err.message}.`) }
          }
        } else {
          log.e("💥💥 unexpected error: ", err)
          this.lastErrorMessage = `${err}`
          return { data: null, err: Error(`💥💥 searchThing error: in Try catch : ${err}`) }
        }
      }
    },
    async create(id: string, t: Thing): Promise<netThing> {
      log.t(`> Entering.. createThing: ${id}`)
      if (t.pos_x !== undefined) t.pos_x = +t.pos_x
      if (t.pos_y !== undefined) t.pos_y = +t.pos_y
      this.clearError()
      this.areWeReady = false
      try {
        const resp = await myAxios.post("thing", t)
        log.l("myAxios.post(thing) : ", resp)
        this.areWeReady = true
        return { data: resp.data, err: null }
      } catch (err) {
        this.areWeReady = true
        if (axios.isAxiosError(err)) {
          log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
          if (err.response !== undefined && err.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
            return { data: null, err: Error(`createThing error : ${err.message}. Server says : ${srvMessage}`) }
          } else {
            return { data: null, err: Error(`createThing error : ${err.message}.`) }
          }
        } else {
          log.e("💥💥 unexpected error: ", err)
          return { data: null, err: Error(`💥💥 createThing error: in deleteThing Try catch : ${err}`) }
        }
      }
    },
    async update(id: string, t: Thing): Promise<netThing> {
      log.t(`> Entering.. updateThing: ${id}`)
      if (t.pos_x !== undefined) t.pos_x = +t.pos_x
      if (t.pos_y !== undefined) t.pos_y = +t.pos_y
      this.clearError()
      this.areWeReady = false
      try {
        const resp = await myAxios.put("thing/" + id, t)
        log.l("myAxios.put(thing/id) : ", resp)
        this.areWeReady = true
        return { data: resp.data, err: null }
      } catch (err) {
        this.areWeReady = true
        if (axios.isAxiosError(err)) {
          log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
          if (err.response !== undefined && err.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
            return { data: null, err: Error(`updateThing error : ${err.message}. Server says : ${srvMessage}`) }
          } else {
            return { data: null, err: Error(`updateThing error : ${err.message}.`) }
          }
        } else {
          log.e("💥💥 unexpected error: ", err)
          return { data: null, err: Error(`💥💥 updateThing error: in deleteThing Try catch : ${err}`) }
        }
      }
    },
    async count(params: ISearchThingParameters): Promise<number> {
      log.t(`> Entering.. typeThing: ${params.typeThing}, createdBy: ${params.createdBy} `)
      this.clearError()
      const urlParams = getUrlParameters(params)
      try {
        //myApi.count(keywords, typeThing, createdBy, myProps.inactivated, myProps.validated)
        const resp = await myAxios.get("thing/search" + urlParams)
        return resp.data
      } catch (error) {
        if (axios.isAxiosError(error)) {
          if (error.response !== undefined) {
            log.w(`countThing error : ${error.message}. Server says : ${error.response.data}`)
          } else {
            log.w(`countThing error : ${error.message}.`)
          }
        } else {
          log.e("unexpected error: ", error)
        }
        return 0
      }
    },
    async delete(id: string): Promise<netThing> {
      log.t(`> Entering.. deleteThing: ${id}`)
      this.clearError()
      this.areWeReady = false
      try {
        const resp = await myAxios.delete("thing/" + id)
        log.l("myAPi._delete : ", resp)
        this.areWeReady = true
        return { data: null, err: null }
      } catch (err) {
        this.areWeReady = true
        if (axios.isAxiosError(err)) {
          log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
          if (err.response !== undefined && err.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
            return { data: null, err: Error(`deleteThing error : ${err.message}. Server says : ${srvMessage}`) }
          } else {
            return { data: null, err: Error(`deleteThing error : ${err.message}.`) }
          }
        } else {
          log.e("💥💥 unexpected error: ", err)
          return { data: null, err: Error(`💥💥 unexpected error: in deleteThing Try catch : ${err}`) }
        }
      }
    },
    async getTypes(): Promise<netThing> {
      log.t(`> Entering getTypes:`)
      this.clearError()
      this.areWeReady = false
      try {
        const resp = await myAxios.get("types")
        log.l(`SUCCESS myAPi.getTypes`)
        if (resp.status == 200) {
          resp.data.forEach((r: TypeThingList) => {
            this.arrListTypeThing.push(r)
          })
          this.areWeReady = true
          return { data: resp.data, err: null }
        } else {
          this.areWeReady = true
          log.w("getTypes got problem", resp)
          return { data: null, err: Error(`problem in getTypes status : ${resp.status}, ${resp.statusText}`) }
        }
      } catch (error) {
        if (axios.isAxiosError(error)) {
          log.w(`getTypes Try Catch Axios ERROR message:${error.message}, error:`, error)
          if (error.response !== undefined && error.response.data !== undefined) {
            const srvMessage = isNullOrUndefined(error.response.data.message) ? "" : error.response.data.message
            const msg = `getTypes error : ${error.message}. Server says : ${srvMessage}`
            log.w(msg)
            return { data: null, err: Error(msg) }
          } else {
            return { data: null, err: Error(`getTypes error : ${error.message}.`) }
          }
        } else {
          log.e("unexpected error: ", error)
          return { data: null, err: Error(`unexpected error: in getTypes Try catch : ${error}`) }
        }
      }
    },
    getIconPath(idType: number): string {
      this.arrListTypeThing.forEach((e) => {
        if (e.id == idType) return e.icon_path
      })
      return "/img/gomarker_star_blue.png"
    },
    async init(searchParams: ISearchThingParameters) {
      log.t(`> Entering ThingStore init`)
      this.clearError()
      this.areWeReady = false
      this.searchParameters = Object.assign({}, searchParams)
      myAxios = axios.create({
        baseURL: BACKEND_URL + "/goapi/v1",
        timeout: defaultAxiosTimeout,
        headers: { "X-Goeland-Token": getSessionId(), Authorization: `Bearer ${getLocalJwtTokenAuth()}` },
      } as CreateAxiosDefaults)
      const res = await this.getTypes()
      if (res.err === null) {
        log.l(`ok, doing getTypes() `)
      } else {
        const msg = `problem in ThingStore init doing getTypes() error:${res.err.message}`
        log.w(msg)
      }
      this.dicoTypeThing = Object.fromEntries(this.arrListTypeThing.map((x) => [x.id, x.name]))
      this.dicoTypeThingIconPath = Object.fromEntries(this.arrListTypeThing.map((x) => [x.id, x.icon_path]))
      this.areWeReady = true
      this.isInitDone = true
    },
    clearError(): void {
      this.isThereAnError = false
      this.lastErrorMessage = ""
    },
  },
})

const getUrlParameters = (searchParam: ISearchThingParameters): string => {
  const localSearchParam = Object.assign({}, searchParam)
  if (localSearchParam.limit == undefined) localSearchParam.limit = defaultQueryLimit
  if (localSearchParam.offset == undefined) localSearchParam.offset = 0
  if (localSearchParam.inactivated == undefined) localSearchParam.inactivated = false
  if (searchParam.typeThing != undefined) {
    localSearchParam.typeThing = searchParam.typeThing == 0 ? undefined : searchParam.typeThing
  }
  if (searchParam.searchKeywords != undefined) {
    localSearchParam.searchKeywords = searchParam.searchKeywords == "" ? undefined : searchParam.searchKeywords
  }
  if (searchParam.createdBy != undefined) {
    localSearchParam.createdBy = searchParam.createdBy == 0 ? undefined : searchParam.createdBy
  }
  let urlParams = `?inactivated=${localSearchParam.inactivated}&limit=${localSearchParam.limit}&offset=${localSearchParam.offset}`
  urlParams += localSearchParam.searchKeywords != undefined ? `&keywords=${localSearchParam.searchKeywords}` : ""
  urlParams += localSearchParam.typeThing != undefined ? `&type=${localSearchParam.typeThing}` : ""
  urlParams += localSearchParam.createdBy != undefined ? `&created_by=${localSearchParam.createdBy}` : ""
  urlParams += localSearchParam.validated !== undefined ? `&validated=${localSearchParam.validated}` : ""
  log.t(
    `After adjusting typeThing: ${localSearchParam.typeThing}, keywords: ${localSearchParam.searchKeywords}, createdBy: ${localSearchParam.createdBy} `
  )
  return urlParams
}
