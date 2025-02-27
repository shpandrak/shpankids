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
 * @interface ApiProblemAnswerForEdit
 */
export interface ApiProblemAnswerForEdit {
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswerForEdit
     */
    id?: string;
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswerForEdit
     */
    title: string;
    /**
     * 
     * @type {boolean}
     * @memberof ApiProblemAnswerForEdit
     */
    isCorrect: boolean;
    /**
     * 
     * @type {string}
     * @memberof ApiProblemAnswerForEdit
     */
    description?: string;
}

/**
 * Check if a given object implements the ApiProblemAnswerForEdit interface.
 */
export function instanceOfApiProblemAnswerForEdit(value: object): boolean {
    if (!('title' in value)) return false;
    if (!('isCorrect' in value)) return false;
    return true;
}

export function ApiProblemAnswerForEditFromJSON(json: any): ApiProblemAnswerForEdit {
    return ApiProblemAnswerForEditFromJSONTyped(json, false);
}

export function ApiProblemAnswerForEditFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiProblemAnswerForEdit {
    if (json == null) {
        return json;
    }
    return {
        
        'id': json['id'] == null ? undefined : json['id'],
        'title': json['title'],
        'isCorrect': json['isCorrect'],
        'description': json['description'] == null ? undefined : json['description'],
    };
}

export function ApiProblemAnswerForEditToJSON(value?: ApiProblemAnswerForEdit | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'id': value['id'],
        'title': value['title'],
        'isCorrect': value['isCorrect'],
        'description': value['description'],
    };
}

