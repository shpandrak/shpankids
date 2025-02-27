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
import type { ApiProblemForEdit } from './ApiProblemForEdit';
import {
    ApiProblemForEditFromJSON,
    ApiProblemForEditFromJSONTyped,
    ApiProblemForEditToJSON,
} from './ApiProblemForEdit';

/**
 * 
 * @export
 * @interface ApiCreateProblemsInSetCommandArgs
 */
export interface ApiCreateProblemsInSetCommandArgs {
    /**
     * 
     * @type {string}
     * @memberof ApiCreateProblemsInSetCommandArgs
     */
    problemSetId: string;
    /**
     * 
     * @type {string}
     * @memberof ApiCreateProblemsInSetCommandArgs
     */
    forUserId: string;
    /**
     * 
     * @type {Array<ApiProblemForEdit>}
     * @memberof ApiCreateProblemsInSetCommandArgs
     */
    problems: Array<ApiProblemForEdit>;
}

/**
 * Check if a given object implements the ApiCreateProblemsInSetCommandArgs interface.
 */
export function instanceOfApiCreateProblemsInSetCommandArgs(value: object): boolean {
    if (!('problemSetId' in value)) return false;
    if (!('forUserId' in value)) return false;
    if (!('problems' in value)) return false;
    return true;
}

export function ApiCreateProblemsInSetCommandArgsFromJSON(json: any): ApiCreateProblemsInSetCommandArgs {
    return ApiCreateProblemsInSetCommandArgsFromJSONTyped(json, false);
}

export function ApiCreateProblemsInSetCommandArgsFromJSONTyped(json: any, ignoreDiscriminator: boolean): ApiCreateProblemsInSetCommandArgs {
    if (json == null) {
        return json;
    }
    return {
        
        'problemSetId': json['problemSetId'],
        'forUserId': json['forUserId'],
        'problems': ((json['problems'] as Array<any>).map(ApiProblemForEditFromJSON)),
    };
}

export function ApiCreateProblemsInSetCommandArgsToJSON(value?: ApiCreateProblemsInSetCommandArgs | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'problemSetId': value['problemSetId'],
        'forUserId': value['forUserId'],
        'problems': ((value['problems'] as Array<any>).map(ApiProblemForEditToJSON)),
    };
}

