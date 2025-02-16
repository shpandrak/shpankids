import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi, uiApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiFamilyTask, UIFamilyInfo} from "../../openapi";
import FamilyTaskEditor from "./FamilyTaskEditor.tsx";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faTrash} from "@fortawesome/free-solid-svg-icons";


export interface FamilyPageProps {
    uiCtx: UiCtx;
}

const FamilyPage: React.FC<FamilyPageProps> = (props) => {

    const [familyInfo, setFamilyInfo] = React.useState<UIFamilyInfo>();
    const [loading, setLoading] = React.useState(true);
    const [subComponent, setSubComponent] = React.useState<React.JSX.Element>();

    React.useEffect(() => {
        uiApi.getFamilyInfo()
            .then(setFamilyInfo)
            .then(() => setLoading(false))
            .catch(showError);

    }, []);

    // create map of family members by email
    const familyMembersByEmail = new Map<string, UIFamilyInfo["members"][0]>();
    if (familyInfo) {
        familyInfo!.members.forEach((member) => {
            familyMembersByEmail.set(member.email, member);
        });
    }

    return (
        <div>
            <h2>Family</h2>
            {loading && <div>Loading...</div>}
            {familyInfo && <div>
                <h3>Family members</h3>
                <ul>
                    {familyInfo.members.map((member) => (
                        <li key={member.email}>{member.firstName} {member.lastName} ({member.role})</li>
                    ))}
                </ul>

                <h3>Family tasks</h3>
                <ul>
                    {familyInfo.tasks
                        .sort((a, b) => a.id.localeCompare(b.id))
                        .map((task) => (
                            <>
                                <li key={task.id}>
                                    {task.title}&nbsp;
                                    ({task.memberIds
                                    .sort()
                                    .map((memberId) => familyMembersByEmail.get(memberId)?.firstName)
                                    .join(", ")
                                })

                                    <button onClick={() => {
                                        const result = window.confirm("Are you sure you want to delete task '" + task.title + "'?");
                                        if (result) {
                                            shpanKidsApi.deleteFamilyTask({apiDeleteFamilyTaskCommandArgs: {taskId: task.id}})
                                                .then(() => {
                                                    uiApi.getFamilyInfo()
                                                        .then(setFamilyInfo)
                                                })
                                                .catch(showError)
                                        }

                                    }}><FontAwesomeIcon title={"Make this task go away"} icon={faTrash}/>
                                    </button>
                                </li>
                            </>
                        ))}
                </ul>

                <button onClick={() => {
                    setSubComponent(
                        <FamilyTaskEditor
                            uiCtx={props.uiCtx}
                            familyTask={{
                                title: "",
                                description: "",
                                memberIds: []
                            }}
                            familyInfo={familyInfo}
                            buttonLabel="Add Task"
                            onSubmit={(task: ApiFamilyTask) => {
                                shpanKidsApi.createFamilyTask({apiCreateFamilyTaskCommandArgs: {task: task}})
                                    .then(() => {
                                        uiApi.getFamilyInfo()
                                            .then(setFamilyInfo)
                                    })
                                    .then(() => setSubComponent(undefined))
                                    .catch(showError)
                            }}
                        />
                    )
                }}>Add Task
                </button>

            </div>}
            {subComponent}
        </div>
    );

}
export default FamilyPage;