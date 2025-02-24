import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {showError} from "../Util.ts";
import {ApiProblem, ApiProblemSet, ApiUserProblemSolution} from "../../openapi";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faInfo} from "@fortawesome/free-solid-svg-icons";


export interface SolutionsAnalysisCompProps {
    uiCtx: UiCtx;
    solutions: ApiUserProblemSolution[];
    problemSet: ApiProblemSet;
    problemFetcher: (problemId: string) => Promise<ApiProblem>;
}

const SolutionsAnalysisComp: React.FC<SolutionsAnalysisCompProps> = (props) => {


    const incorrectSolutions = props.solutions.filter(sol => !sol.correct);
    const correctSolutions = props.solutions.filter(sol => !sol.correct);
    return (
        <>
            <h2>Problem solution analysis for problem set {props.problemSet.title}</h2>
            <i>See side by side view of correct and incorrect solutions</i>
            <hr/>
            <table width={"100%"}>
                <tbody>
                <tr>
                    <td>
                        {incorrectSolutions.length === 0 ? <h3>No incorrect solutions</h3> :
                            (
                                <>
                                    <h3>{incorrectSolutions.length} Incorrect solutions</h3>
                                    {incorrectSolutions.map(sol => {
                                        return (
                                            <td>
                                                <table width={"100%"}>
                                                    <tbody>
                                                    <tr key={sol.problemId}>
                                                        <td>
                                                            {sol.problemTitle}
                                                        </td>
                                                        <td>
                                                            <button onClick={() => {
                                                                props.problemFetcher(sol.problemId)
                                                                    .then(problem => {
                                                                        //todo:
                                                                        alert(JSON.stringify(problem, null, 2));
                                                                    })
                                                                    .catch(showError);
                                                            }}><FontAwesomeIcon icon={faInfo}/></button>
                                                        </td>

                                                    </tr>
                                                    </tbody>
                                                </table>
                                            </td>
                                        )
                                    })}
                                </>
                            )
                        }
                    </td>
                    <td>
                        {correctSolutions.length === 0 ? <h3>No correct solutions</h3> :
                            (
                                <>
                                    <h3>{correctSolutions.length} Correct solutions</h3>
                                    {correctSolutions.map(sol => {
                                        return (
                                            <td>
                                                <table width={"100%"}>
                                                    <tbody>
                                                    <tr key={sol.problemId}>
                                                        <td>
                                                            {sol.problemTitle}
                                                        </td>
                                                        <td>
                                                            <button onClick={() => {
                                                                props.problemFetcher(sol.problemId)
                                                                    .then(problem => {
                                                                        //todo:
                                                                        alert(JSON.stringify(problem, null, 2));
                                                                    })
                                                                    .catch(showError);
                                                            }}><FontAwesomeIcon icon={faInfo}/></button>
                                                        </td>
                                                    </tr>
                                                    </tbody>
                                                </table>
                                            </td>
                                        )
                                    })}
                                </>
                            )
                        }
                    </td>
                </tr>
                </tbody>
            </table>
        </>
    );

}
export default SolutionsAnalysisComp;