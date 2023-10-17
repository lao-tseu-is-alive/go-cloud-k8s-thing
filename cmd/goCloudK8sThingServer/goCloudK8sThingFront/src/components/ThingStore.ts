import { getLog, BACKEND_URL, defaultAxiosTimeout } from "@/config"
import { isNullOrUndefined } from "@/tools/utils"
import { Thing, ThingList } from "@/openapi-generator-cli_thing_typescript-axios"
import axios, { AxiosInstance, CreateAxiosDefaults } from "axios"
import { ref } from "vue/dist/vue"
import { getLocalJwtTokenAuth, getSessionId } from "@/components/Login"

const log = getLog("ThingStore", 4, 2)
let myAxios: AxiosInstance
const areWeReady = ref(false)
const numThingsFound = ref(0)
const searchLimit = ref(10)

// const searchParams: ISearchThingParameters = {}

export interface ISearchThingParameters {
  typeThing?: number | undefined
  searchKeywords?: string | undefined
  createdBy?: number | undefined
  inactivated?: boolean | undefined
  validated?: boolean | undefined
  limit?: number | undefined
  offset?: number | undefined
}
interface IDictionary {
  [key: number]: string
}
const dicoTypeThing: IDictionary = {}
const records: ThingList[] = []
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
export const clearRecords = (): void => {
  if (records.length > 0) {
    records.splice(0)
  }
}

type netThing = { data: Thing | null; err: Error | null }

export const getThing = async (id: string): Promise<netThing> => {
  log.t(`> Entering.. getThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.get("thing/" + id)
    log.l(`SUCCESS myAPi.get(id:${resp.data.id}`)
    if (resp.status == 200) {
      areWeReady.value = true
      return { data: resp.data, err: null }
    } else {
      areWeReady.value = true
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
}

export const createThing = async (id: string, t: Thing): Promise<netThing> => {
  log.t(`> Entering.. createThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.post("thing", t)
    log.l("myAxios.post(thing) : ", resp)
    areWeReady.value = true
    return { data: resp.data, err: null }
  } catch (err) {
    areWeReady.value = true
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
}
export const updateThing = async (id: string, t: Thing): Promise<netThing> => {
  log.t(`> Entering.. updateThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.put("thing/" + id, t)
    log.l("myAxios.put(thing/id) : ", resp)
    areWeReady.value = true
    return { data: resp.data, err: null }
  } catch (err) {
    areWeReady.value = true
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
}
export const searchThing = async (typeThing?: number, keywords?: string, createdBy?: number) => {
  log.t(`> Entering.. typeThing: ${typeThing}, createdBy: ${createdBy} `)
  areWeReady.value = false
  const urlParams = getUrlParameters({ typeThing: typeThing, searchKeywords: keywords, createdBy: createdBy })
  try {
    const resp = await myAxios.get("thing/search" + urlParams)
    log.l("myAxios.get(thing/search) : ", resp)
    clearRecords()
    resp.data.forEach((r: Thing) => {
      records.push(r)
    })
    numThingsFound.value = await countThing(keywords, typeThing, createdBy)
    areWeReady.value = true
    return { data: resp.data, err: null }
  } catch (err) {
    clearRecords()
    numThingsFound.value = await countThing(keywords, typeThing, createdBy)
    areWeReady.value = true
    if (axios.isAxiosError(err)) {
      log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
      if (err.response !== undefined && err.response.data !== undefined) {
        const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
        return { data: null, err: Error(`searchThing error : ${err.message}. Server says : ${srvMessage}`) }
      } else {
        return { data: null, err: Error(`searchThing error : ${err.message}.`) }
      }
    } else {
      log.e("💥💥 unexpected error: ", err)
      return { data: null, err: Error(`💥💥 searchThing error: in Try catch : ${err}`) }
    }
  }
}

export const countThing = async (keywords?: string, typeThing?: number, createdBy?: number): Promise<number> => {
  log.t(`> Entering.. typeThing: ${typeThing}, createdBy: ${createdBy} `)
  if (typeThing != undefined) {
    typeThing = typeThing == 0 ? undefined : typeThing
  }
  if (createdBy != undefined) {
    createdBy = createdBy == 0 ? undefined : createdBy
  }
  log.t(`After adjusting typeThing: ${typeThing}, createdBy: ${createdBy} `)
  const urlParams = getUrlParameters({ typeThing: typeThing, searchKeywords: keywords, createdBy: createdBy })
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
}

export const deleteThing = async (id: string): Promise<netThing> => {
  log.t(`> Entering.. deleteThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.delete("thing/" + id)
    log.l("myAPi._delete : ", resp)
    areWeReady.value = true
    return { data: null, err: null }
  } catch (err) {
    areWeReady.value = true
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
}

export const getTypeThingName = (id: number): string => {
  if (id in dicoTypeThing) {
    return dicoTypeThing[id]
  }
  return "# type inconnu #"
}

export const initStore = async () => {
  myAxios = axios.create({
    baseURL: BACKEND_URL + "/goapi/v1",
    timeout: defaultAxiosTimeout,
    headers: { "X-Goeland-Token": getSessionId(), Authorization: `Bearer ${getLocalJwtTokenAuth()}` },
  } as CreateAxiosDefaults)
  areWeReady.value = true
}

const getUrlParameters = (searchParam: ISearchThingParameters): string => {
  const localSearchParam = Object.assign({}, searchParam)
  if (searchParam.typeThing != undefined) {
    localSearchParam.typeThing = searchParam.typeThing == 0 ? undefined : searchParam.typeThing
  }
  if (searchParam.searchKeywords != undefined) {
    localSearchParam.searchKeywords = searchParam.searchKeywords == "" ? undefined : searchParam.searchKeywords
  }
  if (searchParam.createdBy != undefined) {
    localSearchParam.createdBy = searchParam.createdBy == 0 ? undefined : searchParam.createdBy
  }
  let urlParams = `?inactivated=${searchParam.inactivated}&limit=${searchLimit.value}&offset=${localSearchParam.offset}`
  urlParams += localSearchParam.searchKeywords != undefined ? `&keywords=${localSearchParam.searchKeywords}` : ""
  urlParams += localSearchParam.typeThing != undefined ? `&type=${localSearchParam.typeThing}` : ""
  urlParams += localSearchParam.createdBy != undefined ? `&created_by=${localSearchParam.createdBy}` : ""
  urlParams += localSearchParam.validated !== undefined ? `&validated=${localSearchParam.validated}` : ""
  log.t(
    `After adjusting typeThing: ${localSearchParam.typeThing}, keywords: ${localSearchParam.searchKeywords}, createdBy: ${localSearchParam.createdBy} `
  )
  return urlParams
}
