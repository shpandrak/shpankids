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
                //radio group for answers
                {props.problem.answers.map((answer) => (
                    <div key={answer.id}>
                        <input type="radio"
                               id={answer.id}
                               name="answer"
                               value={answer.id}
                               onChange={(e) => setSelectedAnswerId(e.target.value)}
                        />
                        <label htmlFor={answer.id}>{answer.title}</label>
                    </div>
                ))}
            </div>
        </div>
    );
}
export default ProblemComponent;