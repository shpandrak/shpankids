import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {ApiProblemSet} from "../../openapi";


export interface ProblemSetDetailsEditorProps {
    uiCtx: UiCtx;
    onChange: (problemSet: ApiProblemSet) => void;
    problemSet: ApiProblemSet;
}

const ProblemSetDetailsEditor: React.FC<ProblemSetDetailsEditorProps> = (props) => {
    return (
        <div>
            <div style={{display: "grid", gridTemplateColumns: "1fr 3fr", gap: "10px", textAlign: "left"}}>
                <label>Title</label>
                <input type="text" value={props.problemSet.title} onChange={
                    (e) => props.onChange({...props.problemSet, title: e.target.value})
                }/>
                <label>Description</label>
                <input type="text" value={props.problemSet.description} onChange={
                    (e) => props.onChange({...props.problemSet, description: e.target.value})
                }/>
            </div>
        </div>
    );

}
export default ProblemSetDetailsEditor;