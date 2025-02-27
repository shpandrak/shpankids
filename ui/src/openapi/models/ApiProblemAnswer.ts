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
/**
 * 
 * @export
 * @interface ApiProblemAnswer
 */
export interface ApiProblemAnswer {
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswer
     */
    id: string;
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswer
     */
    title: string;
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswer
     */
    description?: string;
}

/**
 * Check if a given object implements the ApiProblemAnswer interface.
 */
export function instanceOfApiProblemAnswer(value: object): boolean {
    if (!('id' in value)) return false;
    if (!('title' in value)) return false;
    return true;
}

export function ApiProblemAnswerFromJSON(json: any): ApiProblemAnswer {
    return ApiProblemAnswerFromJSONTyped(json, false);
}

export function ApiProblemAnswerFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiProblemAnswer {
    if (json == null) {
        return json;
    }
    return {
        
        'id': json['id'],
        'title': json['title'],
        'description': json['description'] == null ? undefined : json['description'],
    };
}

export function ApiProblemAnswerToJSON(value?: ApiProblemAnswer | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'id': value['id'],
        'title': value['title'],
        'description': value['description'],
    };
}

