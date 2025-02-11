/* tslint:disable */
/* eslint-disable */
/**
 * ShpanKids API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 0.1
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { mapValues } from '../runtime';
import type { ApiTaskStatus } from './ApiTaskStatus';
import {
    ApiTaskStatusFromJSON,
    ApiTaskStatusFromJSONTyped,
    ApiTaskStatusToJSON,
} from './ApiTaskStatus';

/**
 * Task
 * @export
 * @interface ApiTask
 */
export interface ApiTask {
    /**
     * 
     * @type {string}
     * @memberof ApiTask
     */
    id: string;
    /**
     * 
     * @type {string}
     * @memberof ApiTask
     */
    title: string;
    /**
     * 
     * @type {string}
     * @memberof ApiTask
     */
    description: string;
    /**
     * 
     * @type {Date}
     * @memberof ApiTask
     */
    forDate: Date;
    /**
     * 
     * @type {Date}
     * @memberof ApiTask
     */
    dueDate?: Date;
    /**
     * 
     * @type {ApiTaskStatus}
     * @memberof ApiTask
     */
    status: ApiTaskStatus;
}

/**
 * Check if a given object implements the ApiTask interface.
 */
export function instanceOfApiTask(value: object): boolean {
    if (!('id' in value)) return false;
    if (!('title' in value)) return false;
    if (!('description' in value)) return false;
    if (!('forDate' in value)) return false;
    if (!('status' in value)) return false;
    return true;
}

export function ApiTaskFromJSON(json: any): ApiTask {
    return ApiTaskFromJSONTyped(json, false);
}

export function ApiTaskFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiTask {
    if (json == null) {
        return json;
    }
    return {
        
        'id': json['id'],
        'title': json['title'],
        'description': json['description'],
        'forDate': (new Date(json['forDate'])),
        'dueDate': json['dueDate'] == null ? undefined : (new Date(json['dueDate'])),
        'status': ApiTaskStatusFromJSON(json['status']),
    };
}

export function ApiTaskToJSON(value?: ApiTask | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'id': value['id'],
        'title': value['title'],
        'description': value['description'],
        'forDate': ((value['forDate']).toISOString()),
        'dueDate': value['dueDate'] == null ? undefined : ((value['dueDate']).toISOString()),
        'status': ApiTaskStatusToJSON(value['status']),
    };
}

