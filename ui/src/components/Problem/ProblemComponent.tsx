import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {ApiProblem} from "../../openapi";


export interface ProblemComponentProps {
    uiCtx: UiCtx;
    problem: ApiProblem;
}

const ProblemComponent: React.FC<ProblemComponentProps> = (props) => {

    const [selectedAnswerId, setSelectedAnswerId] = React.useState<string>();

    // create map of family members by email
    return (
        <div>
            <h3>Problem</h3>
            <div>
                {props.problem.title}
            </div>
            {props.problem.description && (<div>{props.problem.description}</div>)}
            <div>
                <h3>Answers</h3>
                <table>
                    <tbody>

                        {props.problem.answers.map((answer) => (
                        <tr key={answer.id}>
                            <td>
                                <input type="radio"
                                       id={answer.id}
                                       name="answer"
                                       value={answer.id}
                                       onChange={(e) => setSelectedAnswerId(e.target.value)}
                                />
                            </td>
                            <td>
                                <label htmlFor={answer.id}>{answer.title}</label>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>

            </div>
        </div>
    );
}
export default ProblemComponent;