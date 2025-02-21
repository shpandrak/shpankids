import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiProblemForEdit, ApiProblemSet, UIFamilyInfo, UIFamilyMember} from "../../openapi";
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {faList, faTrash} from "@fortawesome/free-solid-svg-icons";
import ProblemSetEditor from "./ProblemSetEditor.tsx";
import ProblemSetDetailsEditor from "./ProblemSetDetailsEditor.tsx";


export interface ProblemSetsPageProps {
    uiCtx: UiCtx;
    familyInfo: UIFamilyInfo;
}

const newPs = {
    id: "",
    title: "New problem set",
    description: ""
};
const ProblemSetsPage: React.FC<ProblemSetsPageProps> = (props) => {


    const [familyMember, setSelectedFamilyMember] = React.useState<UIFamilyMember>(props.familyInfo.members[0]);
    const [loading, setLoading] = React.useState(true);
    const [problemSets, setProblemSets] = React.useState<ApiProblemSet[]>();
    const [subComponent, setSubComponent] = React.useState<React.JSX.Element>();
    const [newProblemSet, setNewProblemSet] = React.useState<ApiProblemSet>();

    function reloadProblemSets() {
        setNewProblemSet(undefined);
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
    }

    React.useEffect(() => {
        reloadProblemSets();
    }, [familyMember]);

    function reloadProblemsList(problemSet: ApiProblemSet) {
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
                            return shpanKidsApi.createProblemsInSet(
                                {
                                    apiCreateProblemsInSetCommandArgs: {
                                        problemSetId: problemSet.id,
                                        problems: problemsToCreate,
                                        forUserId: familyMember.email
                                    }
                                }
                            ).then(() => {
                                reloadProblemsList(problemSet);
                            });
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
                        refineProblemsHandler={(
                            problemSetId: string,
                            userId: string,
                            refineText: string,
                            problemsToRefine: ApiProblemForEdit[]
                        ): Promise<ApiProblemForEdit[]> => {
                            return shpanKidsApi.refineProblems({
                                apiRefineProblemsCommandArgs: {
                                    problemSetId: problemSetId,
                                    userId: userId,
                                    refineText: refineText,
                                    problems: problemsToRefine
                                }
                            });
                        }}
                        generateProblemsHandler={(
                            problemSetId: string,
                            userId: string,
                            additionalRequestText?: string): Promise<ApiProblemForEdit[]> => {
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
    }

    return (
        <div>
            <h2>Problem sets for {familyMember.firstName}</h2>
            {loading && <div>Loading...</div>}
            {problemSets && <div>
                <h3>Select a family member</h3>
                <div>
                    <select value={familyMember.email} onChange={
                        (e) => {
                            setSubComponent(undefined);
                            setNewProblemSet(undefined);
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
                                    reloadProblemsList(problemSet);
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
                <button onClick={() => {
                    setNewProblemSet(newPs);
                }}>Create New Problem Set
                </button>

            </div>}
            {subComponent}
            {newProblemSet && (
                <div>
                    <h3>
                        Create New Problem Set
                    </h3>
                    <ProblemSetDetailsEditor
                        problemSet={newProblemSet}
                        uiCtx={props.uiCtx}
                        onChange={setNewProblemSet}
                    />
                    <button onClick={() => {
                        shpanKidsApi.createProblemSet({
                            apiCreateProblemSetCommandArgs: {
                                title: newProblemSet?.title ?? "",
                                description: newProblemSet?.description ?? "",
                                forUserId: familyMember.email
                            }
                        })
                            .then(() => {
                                reloadProblemSets()
                            })
                            .catch(showError);
                    }}>
                        Create
                    </button>
                </div>
            )}
        </div>
    );

}
export default ProblemSetsPage;