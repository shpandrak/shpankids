import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {ApiFamilyTask, UIFamilyInfo} from "../../openapi";


export interface FamilyTaskEditorProps {
    uiCtx: UiCtx;
    familyTask: ApiFamilyTask;
    familyInfo: UIFamilyInfo;
    buttonLabel: string;
    onSubmit: (task: ApiFamilyTask) => void;
}

const FamilyTaskEditor: React.FC<FamilyTaskEditorProps> = (props) => {

    const [familyTask, setFamilyTask] = React.useState<ApiFamilyTask>(props.familyTask);
    return (
        <div>
            <h2>Family Task</h2>
            <div style={{display: "grid", gridTemplateColumns: "1fr 3fr", gap: "10px", textAlign: "left"}}>
                <label>Title</label>
                <input type="text" value={familyTask.title} onChange={
                    (e) => setFamilyTask({...familyTask, title: e.target.value})
                }/>
                <label>Description</label>
                <input type="text" value={familyTask.description} onChange={
                    (e) => setFamilyTask({...familyTask, description: e.target.value})
                }/>
                <label>Relevant Members</label>
                <select multiple value={familyTask.memberIds} onChange={
                    (e) => setFamilyTask({
                        ...familyTask,
                        memberIds: Array.from(e.target.selectedOptions).map(o => o.value)
                    })
                }>
                    {props.familyInfo.members.map((member) => (
                        <option key={member.email}
                                value={member.email}>{member.firstName} {member.lastName} ({member.role})</option>
                    ))}
                </select>
            </div>
            <button onClick={() => props.onSubmit(familyTask)}>{props.buttonLabel}</button>
        </div>
    );

}
export default FamilyTaskEditor;