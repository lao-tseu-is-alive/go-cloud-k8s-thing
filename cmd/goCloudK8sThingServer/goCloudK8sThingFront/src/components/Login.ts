import { SHA256 } from "crypto-es/lib/sha256"
import axios from "axios"
import { getLog, APP, BACKEND_URL } from "../config"

const log = getLog("Login", 4, 1)

export const getPasswordHash = (password: string) => SHA256(password).toString()

export const parseJwt = (token: string) => {
  const base64Url = token.split(".")[1]
  const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/")
  const jsonPayload = decodeURIComponent(
    atob(base64)
      .split("")
      .map((c) => `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`)
      .join("")
  )

  return JSON.parse(jsonPayload)
}

export const getToken = async (baseServerUrl: string, username: string, passwordHash: string) => {
  const data = {
    username,
    password_hash: `${passwordHash}`,
  }
  log.t("# entering...  data :", data)
  let response = null
  try {
    response = await axios.post(`${baseServerUrl}/login`, data) // .then((response) => {
    log.l("getToken() axios.post Success ! response :", response.data)
    const jwtValues = parseJwt(response.data.token)
    log.l("getToken() token values : ", jwtValues)
    const dExpires = new Date(0)
    dExpires.setUTCSeconds(jwtValues.exp)
    log.l(`getToken() JWT token expiration : ${dExpires}`)
    if (response.status === 200) {
      if (typeof Storage !== "undefined") {
        // Code for localStorage/sessionStorage.
        sessionStorage.setItem(`${APP}_goapi_jwt_session_token`, response.data.token)
        sessionStorage.setItem(`${APP}_goapi_idgouser`, jwtValues.id)
        sessionStorage.setItem(`${APP}_goapi_name`, jwtValues.name)
        sessionStorage.setItem(`${APP}_goapi_username`, username)
        sessionStorage.setItem(`${APP}_goapi_email`, jwtValues.email)
        sessionStorage.setItem(`${APP}_goapi_isadmin`, jwtValues.is_admin)
        sessionStorage.setItem(`${APP}_goapi_groups`, jwtValues.groups)
        sessionStorage.setItem(`${APP}_goapi_date_expiration`, jwtValues.exp)
      }
      return {
        msg: "getToken() axios.post Success.",
        err: null,
        status: response.status,
        data: response.data,
      }
    }
    log.w("axios get a bad status ! response was:", response)
    return {
      msg: `getToken() axios.post Failure got a bad status: ${response.status} !`,
      err: null,
      status: response.status,
      data: null,
    }
  } catch (error) {
    if (axios.isAxiosError(error)) {
      log.w(`Try Catch Axios ERROR message:${error.message}, error:`, error)
      log.l("Axios error.response:", error.response)
      return {
        msg: `getToken() Try Catch Axios ERROR: ${error.message} !`,
        err: error,
        status: error.status,
        data: null,
      }
    } else {
      log.e("unexpected error: ", error)
      return {
        msg: `getToken() Try Catch unexpected ERROR: ${error} !`,
        err: error,
        status: null,
        data: null,
      }
    }
  }
}

export const getTokenStatus = async (baseServerUrl = BACKEND_URL) => {
  log.t("# entering...  ")
  axios.defaults.headers.common.Authorization = `Bearer ${sessionStorage.getItem(`${APP}_goapi_jwt_session_token`)}`
  try {
    const res = await axios.get(`${baseServerUrl}/goapi/v1/status`)
    log.l("getTokenStatus() axios.get Success ! response :", res)
    const dExpires = new Date(0)
    dExpires.setUTCSeconds(res.data.exp)
    const msg = `getTokenStatus() JWT token expiration : ${dExpires}`
    log.l(msg)
    const { data } = res
    return {
      msg,
      err: null,
      status: res.status,
      data,
    }
  } catch (error) {
    if (axios.isAxiosError(error)) {
      log.e("getToken() ## Try Catch ERROR ## error :", error)
      log.e("axios response was:", error.response)
      log.e("axios message is:", error.message)
      const msg = `Error: in getTokenStatus() ## axios.get(${baseServerUrl}/goapi/v1/status) ERROR ## error :${error.message}`
      log.w(msg)
      const status = error.response != undefined ? error.response.status : undefined
      return {
        msg,
        err: error,
        status: status,
        data: null,
      }
    } else {
      const msg = `An unexpected Error occured in getTokenStatus() ## axios.get(${baseServerUrl}/goapi/v1/status) ERROR ## error :${error}`
      log.e(msg)
      log.e("unexpected error: ", error)
      return {
        msg: msg,
        err: error,
        status: null,
        data: null,
      }
    }
  }
}

export const clearSessionStorage = (): void => {
  // Code for localStorage/sessionStorage.
  sessionStorage.removeItem(`${APP}_goapi_jwt_session_token`)
  sessionStorage.removeItem(`${APP}_goapi_idgouser`)
  sessionStorage.removeItem(`${APP}_goapi_name`)
  sessionStorage.removeItem(`${APP}_goapi_username`)
  sessionStorage.removeItem(`${APP}_goapi_email`)
  sessionStorage.removeItem(`${APP}_goapi_isadmin`)
  sessionStorage.removeItem(`${APP}_goapi_groups`)
  sessionStorage.removeItem(`${APP}_goapi_date_expiration`)
}

export const logoutAndResetToken = (baseServerUrl: string) => {
  log.t("# IN logoutAndResetToken()")
  axios.defaults.headers.common.Authorization = `Bearer ${sessionStorage.getItem(`${APP}_goapi_jwt_session_token`)}`
  axios
    .get(`${baseServerUrl}/api/logout`)
    .then((response) => {
      log.l("logoutAndResetToken() axios.get Success ! response :", response)
      clearSessionStorage()
    })
    .catch((error) => {
      log.e("logoutAndResetToken() ## axios.get ERROR ## error :", error)
    })
}

export const doesCurrentSessionExist = (): boolean => {
  log.t("# entering...  ")
  if (sessionStorage.getItem(`${APP}_goapi_jwt_session_token`) == null) return false
  if (sessionStorage.getItem(`${APP}_goapi_idgouser`) == null) return false
  if (sessionStorage.getItem(`${APP}_goapi_isadmin`) == null) return false
  if (sessionStorage.getItem(`${APP}_goapi_email`) == null) return false
  if (sessionStorage.getItem(`${APP}_goapi_date_expiration`) !== null) {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const goapi_token_expires = parseInt(sessionStorage.getItem(`${APP}_goapi_date_expiration`), 10)
    //log.l("goapi_token_expires : ", goapi_token_expires)
    const dateExpire = new Date(goapi_token_expires * 1000)
    //log.l("dateExpire : ", dateExpire)
    const now = new Date()
    const minutesUntilExpire = Math.floor((dateExpire.getTime() - now.getTime()) / 1000 / 60)
    if (now > dateExpire) {
      clearSessionStorage()
      log.w("# IN doesCurrentSessionExist() SESSION EXPIRED")
      return false
    }
    // attention meme si une session existe en local il faut que le jwt token soit  encore valide !
    log.l(`Yes session exists, valid for ${minutesUntilExpire} minutes...`)
    return true
  }
  log.w("# IN doesCurrentSessionExist() goapi_date_expiration was null ")
  return false
}

export const getLocalJwtTokenAuth = (): string => {
  if (doesCurrentSessionExist()) {
    //return `Bearer ${sessionStorage.getItem(`${APP}_goapi_jwt_session_token`)}`
    const goapi_jwt_session = sessionStorage.getItem(`${APP}_goapi_jwt_session_token`)
    if (goapi_jwt_session !== null) {
      return goapi_jwt_session
    }
  }
  return ""
}

export const getDateExpiration = (): number => {
  if (doesCurrentSessionExist()) {
    const val = sessionStorage.getItem(`${APP}_goapi_date_expiration`)
    if (val !== null) {
      return parseInt(val, 10)
    }
  }
  return 0
}

export const getUserEmail = (): string => {
  if (doesCurrentSessionExist()) {
    return `${sessionStorage.getItem(`${APP}_goapi_email`)}`
  }
  return ""
}

export const getUserId = () => {
  if (doesCurrentSessionExist()) {
    return parseInt(`${sessionStorage.getItem(`${APP}_goapi_idgouser`)}`, 10)
  }
  return 0
}

export const getUserName = () => {
  if (doesCurrentSessionExist()) {
    return `${sessionStorage.getItem(`${APP}_goapi_name`)}`
  }
  return ""
}

export const getUserLogin = () => {
  if (doesCurrentSessionExist()) {
    return `${sessionStorage.getItem(`${APP}_goapi_username`)}`
  }
  return ""
}

export const getUserIsAdmin = () => {
  if (doesCurrentSessionExist()) {
    return sessionStorage.getItem(`${APP}_goapi_isadmin`) === "true"
  }
  return false
}

export const getUserFirstGroups = () => {
  if (doesCurrentSessionExist()) {
    if (sessionStorage.getItem(`${APP}_goapi_groups`) == null) return null
    if (sessionStorage.getItem(`${APP}_goapi_groups`) === "null") return null
    // let's clone it and converting to an array of integers
    const tmpArr = sessionStorage.getItem(`${APP}_goapi_groups`)
    if (tmpArr != null) {
      if (tmpArr.indexOf(",") > 0) {
        const firstFiltered = tmpArr.split(",").map((e) => +e)
        return firstFiltered[0]
      }
      return parseInt(tmpArr, 10)
    }
  }
  return null
}

export const getUserGroupsArray = () => {
  if (doesCurrentSessionExist()) {
    if (sessionStorage.getItem(`${APP}_goapi_groups`) == null) return null
    if (sessionStorage.getItem(`${APP}_goapi_groups`) === "null") return null
    // let's clone it and converting to an array of integers
    const tmpArr = sessionStorage.getItem(`${APP}_goapi_groups`)
    if (tmpArr != null) {
      if (tmpArr.indexOf(",") > 0) {
        return tmpArr.split(",").map((i) => parseInt(i, 10))
      }
      return [parseInt(tmpArr, 10)]
    }
  }
  return null
}

export const isUserHavingGroups = () => {
  if (doesCurrentSessionExist()) {
    if (sessionStorage.getItem(`${APP}_goapi_groups`) == null) return false
    if (sessionStorage.getItem(`${APP}_goapi_groups`) === "null") return false
    // let's clone it and converting to an array of integers
    const tmp = sessionStorage.getItem(`${APP}_goapi_groups`)
    if (tmp != null) {
      if (tmp.indexOf(",") > 0) {
        return true
      }
      if (parseInt(tmp, 10) > 0) {
        return true
      }
    }
    return false
  }
  return false
}
