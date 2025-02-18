import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {ApiProblemForEdit, ApiProblemSet} from "../../openapi";
import {showError} from "../Util.ts";
import ProblemEditor from "./ProblemEditor.tsx";


export interface ProblemSetEditorProps {
    uiCtx: UiCtx;
    problems: ApiProblemForEdit[];
    problemSet: ApiProblemSet;
    userId: string;
    generateProblemsHandler: (problemSetId: string, userId: string, additionalRequestText?: string) => Promise<ApiProblemForEdit[]>;
    createNewProblemsHandler: (problemsToCreate: ApiProblemForEdit[]) => Promise<void>;
    deleteProblemHandler: (problemsToEdit: ApiProblemForEdit[]) => Promise<void>;
    updateProblemHandler: (problemSet: ApiProblemForEdit) => Promise<void>;
    updateProblemSetHandler: (problemSet: ApiProblemSet) => Promise<void>;
}

const ProblemSetEditor: React.FC<ProblemSetEditorProps> = (props) => {


    return (
        <div>
            <h2>Edit Problem Set</h2>
            <div style={{display: "grid", gridTemplateColumns: "1fr 3fr", gap: "10px", textAlign: "left"}}>
                <label>Title</label>
                <input type="text" value={props.problemSet.title} onChange={
                    (e) => props.updateProblemSetHandler({...props.problemSet, title: e.target.value})
                }/>
                <label>Description</label>
                <input type="text" value={props.problemSet.description} onChange={
                    (e) => props.updateProblemSetHandler({...props.problemSet, description: e.target.value})
                }/>
            </div>
            <h3>Problems</h3>
            <div>
                <table>
                    <thead>
                    <tr>
                        <th>Title</th>
                        <th>&nbsp;</th>
                    </tr>
                    </thead>
                    <tbody>
                    {props.problems.map((problem, idx) => (
                        <tr key={idx}>
                            <td>{problem.title}</td>
                            <td>{problem.description}</td>
                            <td>
                                <button onClick={() => {
                                    alert("Not implemented yet");
                                }}>Edit
                                </button>
                                <button onClick={() => {
                                    if (window.confirm("Are you sure you want to delete this problem?")) {
                                        props.createNewProblemsHandler([problem])
                                    }
                                }}>Delete
                                </button>

                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
                <button onClick={() => {
                    const additionalText = window.prompt("Enter additional request text")
                    props.generateProblemsHandler(props.problemSet.id, props.userId, additionalText == null ? undefined : additionalText)
                        .then((problems) => {
                            props.uiCtx.showModal((
                                <>
                                    {problems.map((problem, idx) => (
                                        <ProblemEditor
                                            key={idx}
                                            uiCtx={props.uiCtx}
                                            problem={problem}
                                            onChanges={(newProblem) => {
                                                console.log(newProblem)
                                            }
                                            }></ProblemEditor>
                                    ))}
                                </>
                            ))
                        })
                        .catch(showError)
                }}>Generate Next Problems
                </button>
            </div>

        </div>
    );

}
export default ProblemSetEditor;