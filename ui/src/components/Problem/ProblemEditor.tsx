import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faTrash} from "@fortawesome/free-solid-svg-icons";
import {ApiProblemForEdit} from "../../openapi";


export interface ProblemEditorProps {
    uiCtx: UiCtx;
    problem: ApiProblemForEdit;
    onChanges: (problem: ApiProblemForEdit) => void;
}

const ProblemEditor: React.FC<ProblemEditorProps> = (props) => {

    return (
        <div>
            <h2>Edit Problem</h2>
            <div style={{display: "grid", gridTemplateColumns: "1fr 3fr", gap: "10px", textAlign: "left"}}>
                <label>Title</label>
                <textarea
                    dir="auto"
                    value={props.problem.title}
                    onChange={(e) => props.onChanges({ ...props.problem, title: e.target.value })}
                    rows={4} // Adjust the number of rows for initial height
                    style={{ width: '100%' }} // Optional: Make it full width
                />
                <label>Description</label>
                <input type="text" value={props.problem.description} onChange={
                    (e) => props.onChanges({...props.problem, description: e.target.value})
                }/>
                <label>Answers</label>
                <div>

                    <table>
                        <thead>
                        <tr>
                            <th>Correct</th>
                            <th>Answer</th>
                            <th>Description</th>
                            <th>&nbsp;</th>
                        </tr>
                        </thead>
                        <tbody>
                        {props.problem.answers.map((answer, idx) => (
                            <tr key={idx}>
                                <td>
                                    <input type="checkbox" checked={answer.isCorrect} onChange={
                                        (e) => props.onChanges({
                                            ...props.problem,
                                            answers: props.problem.answers.map((a, innerIdx) => {
                                                return innerIdx === idx ? {...a, isCorrect: e.target.checked} : a;
                                            })
                                        })}/>
                                </td>
                                <td>
                                    <input type="text" value={answer.title} onChange={
                                        (e) => props.onChanges({
                                            ...props.problem,
                                            answers: props.problem.answers.map((a, innerIdx) => {
                                                return innerIdx === idx ? {...a, title: e.target.value} : a;
                                            })
                                        })
                                    }/>
                                </td>
                                <td>
                                    <input type="text" value={answer.description} onChange={
                                        (e) => props.onChanges({
                                            ...props.problem,
                                            answers: props.problem.answers.map((a) => {
                                                if (a.id === answer.id) {
                                                    return {...a, description: e.target.value};
                                                }
                                                return a;
                                            })
                                        })
                                    }/>
                                </td>
                                <td>
                                    <button onClick={() => props.onChanges({
                                        ...props.problem,
                                        answers: props.problem.answers.filter((_, idx) => idx !== idx)
                                    })}><FontAwesomeIcon title={"Remove Answer"} icon={faTrash}/></button>
                                </td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                    <button onClick={() => props.onChanges({
                        ...props.problem,
                        answers: [...props.problem.answers, {
                            title: "New Answer " + (props.problem.answers.length + 1),
                            isCorrect: false
                        }]
                    })}>Add Answer
                    </button>
                </div>
            </div>
        </div>
    );

}
export default ProblemEditor;