import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiProblemForEdit, ApiProblemSet, UIFamilyInfo, UIFamilyMember} from "../../openapi";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faList, faTrash} from "@fortawesome/free-solid-svg-icons";
import ProblemSetEditor from "./ProblemSetEditor.tsx";


export interface ProblemSetsPageProps {
    uiCtx: UiCtx;
    familyInfo: UIFamilyInfo;
}

const ProblemSetsPage: React.FC<ProblemSetsPageProps> = (props) => {


    const [familyMember, setSelectedFamilyMember] = React.useState<UIFamilyMember>(props.familyInfo.members[0]);
    const [loading, setLoading] = React.useState(true);
    const [problemSets, setProblemSets] = React.useState<ApiProblemSet[]>();
    const [subComponent, setSubComponent] = React.useState<React.JSX.Element>();

    React.useEffect(() => {
        setLoading(true);
        shpanKidsApi.listUserFamilyProblemSets(
            {
                userId: familyMember.email
            }
        )
            .then(problemSets => problemSets.sort((a, b) => a.id.localeCompare(b.id)))
            .then(setProblemSets)
            .then(() => setLoading(false))
            .catch(showError);


    }, [familyMember]);

    return (
        <div>
            <h2>Problem sets for {familyMember.firstName}</h2>
            {loading && <div>Loading...</div>}
            {problemSets && <div>
                <h3>Select a family member</h3>
                <div>
                    <select value={familyMember.email} onChange={
                        (e) => {
                            setSelectedFamilyMember(props.familyInfo.members.find((member) => member.email === e.target.value)!)
                        }
                    }>
                        {props.familyInfo.members.map((member) => (
                            <option key={member.email}
                                    value={member.email}>{member.firstName} {member.lastName}</option>
                        ))}
                    </select>
                </div>

                <h3>Problem sets</h3>
                <table>
                    <tbody>
                    {problemSets.map((problemSet) => (
                        <tr key={problemSet.id}>
                            <td>{problemSet.title}</td>
                            <td>{problemSet.description}</td>
                            <td>
                                <button onClick={() => {
                                    shpanKidsApi.listProblemSetProblems({
                                        userId: familyMember.email,
                                        problemSetId: problemSet.id
                                    })
                                        .then((problems) => {
                                            setSubComponent(
                                                <ProblemSetEditor
                                                    problemSet={problemSet}
                                                    uiCtx={props.uiCtx}
                                                    problems={problems}
                                                    userId={familyMember.email}
                                                    createNewProblemsHandler={(problemsToCreate): Promise<void> => {
                                                        throw new Error("Not implemented yet");
                                                    }}
                                                    deleteProblemHandler={(problemsToEdit): Promise<void> => {
                                                        throw new Error("Not implemented yet");
                                                    }}
                                                    updateProblemHandler={(problemSet: ApiProblemForEdit): Promise<void> => {
                                                        throw new Error("Not implemented yet");
                                                    }}
                                                    updateProblemSetHandler={(problemSet: ApiProblemSet): Promise<void> => {
                                                        throw new Error("Not implemented yet");
                                                    }}
                                                    generateProblemsHandler={(problemSetId: string, userId: string, additionalRequestText?: string): Promise<ApiProblemForEdit[]> => {
                                                        return shpanKidsApi.generateProblems({
                                                            apiGenerateProblemsCommandArgs: {
                                                                problemSetId: problemSetId,
                                                                userId: userId,
                                                                additionalRequestText: additionalRequestText
                                                            }
                                                        })
                                                    }}

                                                />
                                            );
                                        })
                                        .catch(showError);
                                }}>
                                    <FontAwesomeIcon icon={faList}/>
                                </button>
                                <button onClick={() => {
                                    alert("Not implemented yet");
                                }}>
                                    <FontAwesomeIcon icon={faTrash}/>
                                </button>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>}
            {subComponent}
        </div>
    );

}
export default ProblemSetsPage;