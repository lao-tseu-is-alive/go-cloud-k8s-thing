/**
 * Thing microservice written in Golang
 * OpenApi Specification for an API to manage Thing
 *
 * OpenAPI spec version: 0.0.5
 * Contact: go-cloud-k8s-thing@goeland.io
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 * Do not edit the class manually.
 */
import { ThingStatus } from "./thing-status"
/**
 *
 * @export
 * @interface ThingList
 */
export interface ThingList {
  /**
   *
   * @type {string}
   * @memberof ThingList
   */
  id: string
  /**
   *
   * @type {number}
   * @memberof ThingList
   */
  typeId: number
  /**
   *
   * @type {string}
   * @memberof ThingList
   */
  name: string
  /**
   *
   * @type {string}
   * @memberof ThingList
   */
  description?: string
  /**
   *
   * @type {number}
   * @memberof ThingList
   */
  externalId?: number
  /**
   *
   * @type {boolean}
   * @memberof ThingList
   */
  inactivated: boolean
  /**
   *
   * @type {boolean}
   * @memberof ThingList
   */
  validated?: boolean
  /**
   *
   * @type {ThingStatus}
   * @memberof ThingList
   */
  status?: ThingStatus
  /**
   *
   * @type {number}
   * @memberof ThingList
   */
  createdBy?: number
  /**
   *
   * @type {Date}
   * @memberof ThingList
   */
  createdAt?: Date
  /**
   *
   * @type {number}
   * @memberof ThingList
   */
  posX: number
  /**
   *
   * @type {number}
   * @memberof ThingList
   */
  posY: number
}
