import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiProblemForEdit, ApiProblemSet, UIFamilyInfo, UIFamilyMember} from "../../openapi";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faList, faTrash} from "@fortawesome/free-solid-svg-icons";
import ProblemSetEditor from "./ProblemSetEditor.tsx";
import ProblemEditor from "./ProblemEditor.tsx";


export interface ProblemsSelectorCompProps {
    uiCtx: UiCtx;
    problems: SelectableProblem[];
    updateProblems: (problems: SelectableProblem[]) => void;
}

export class SelectableProblem {
    constructor(
        public problem: ApiProblemForEdit,
        public selected: boolean
    ) {
    }
}

const ProblemsSelectorComp: React.FC<ProblemsSelectorCompProps> = (props) => {

    return (
        <>
            <h2>Suggested Problems</h2>
            <i>Update and select the problems to add to the problem-set</i>
            <hr/>
            <table>
                <tbody>
                {props.problems.map((problem, idx) => (
                    <tr key={idx}>
                        <td>
                            <input type="checkbox" checked={problem.selected} onChange={
                                (e) => {
                                    const newProblems = props.problems.map((p, innerIdx) => {
                                        return innerIdx === idx ? new SelectableProblem(p.problem, e.target.checked) : p;
                                    });
                                    props.updateProblems(newProblems);
                                }
                            }/>
                        </td>
                        <td>
                            <ProblemEditor
                                uiCtx={props.uiCtx}
                                problem={problem.problem}
                                onChanges={(newProblem) => {
                                    const newProblems = props.problems.map((p, innerIdx) => {
                                        return innerIdx === idx ? new SelectableProblem(newProblem, p.selected) : p;
                                    });
                                    props.updateProblems(newProblems);
                                }}
                            />
                        </td>
                    </tr>
                ))}
                </tbody>
            </table>
        </>
    );

}
export default ProblemsSelectorComp;