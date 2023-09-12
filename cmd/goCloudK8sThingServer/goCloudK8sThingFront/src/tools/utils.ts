import { getLog } from "@/config"

const log = getLog("Utils", 2, 2)
/**
 * check if the given variable is null or undefined
 * @param variable
 * @returns true if given variable is null or undefined, false in other cases
 */
export const isNullOrUndefined = (variable: any): boolean => typeof variable === "undefined" || variable === null

/**
 * check if the given variable is null, undefined or an empty string(zero length)
 * @param variable
 * @returns true if given variable is null or undefined or an empty string, false in other cases
 */
export const isEmpty = (variable: any): boolean =>
  typeof variable === "undefined" || variable === null || variable === ""

/**
 * convert a date string from iso yyyy-mm-dd in french europe dd-mm-yyyy
 * @param strIsoDate string yyyy-mm-dd
 * @returns date as string in format dd-mm-yyyy
 */
export const dateIso2Fr = function (strIsoDate: string): string {
  if (isEmpty(strIsoDate)) {
    return ""
  }
  const [y, m, d] = strIsoDate.split("-")
  return [d, m, y].join("-")
}

export const getDateFromTimeStamp = (isoDate: string) => {
  log.t(`#> entering : ${isoDate}`, isoDate)
  if (typeof isoDate !== "string") return "not_date"
  if (isNullOrUndefined(isoDate) || isoDate.indexOf("T") < 0) return ""
  const dateTS = dateIso2Fr(isoDate.split("T")[0])
  log.l(`dateTS : ${dateTS}`, isoDate)
  return dateTS
}

export const truncateText = (text: string, maxSize = 40): string => {
  if (isNullOrUndefined(text)) return ""
  if (text.length < maxSize) return text
  return `${text.substring(0, maxSize)}…`
}
