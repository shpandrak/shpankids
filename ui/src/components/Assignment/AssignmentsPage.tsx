import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiAssignment} from "../../openapi";
import ProblemComponent from "../Problem/ProblemComponent.tsx";


export interface AssignmentsPageProps {
    uiCtx: UiCtx;
}

const AssignmentsPage: React.FC<AssignmentsPageProps> = (props) => {
    const [assignments, setAssignments] = React.useState<ApiAssignment[]>([]);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        shpanKidsApi.listAssignments()
            .then(assignments => assignments.sort((a, b) => a.id.localeCompare(b.id)))
            .then(setAssignments)
            .then(() => setLoading(false))
            .catch(showError);

    }, [props.uiCtx]);

    return (
        <div>
            <h2>Today's tasks</h2>
            {loading && <div>Loading...</div>}
            <ol>
                {assignments.map((assignment) => (
                    <li style={{textDecoration: assignment.status == "done" ? "line-through" : "auto"}}
                        key={assignment.id}>{assignment.title}
                        {assignment.type == "task" && assignment.status == "open" && (
                            <button onClick={() => {
                                shpanKidsApi.updateTaskStatus({
                                    apiUpdateTaskStatusCommandArgs: {
                                        taskId: assignment.id,
                                        status: "done",
                                        forDate: assignment.forDate
                                    }
                                })
                                    .then(() => {
                                        setAssignments(assignments.map((t) => {
                                            if (t.id === assignment.id) {
                                                return {...t, status: "done"};
                                            }
                                            return t;
                                        }));
                                    })
                                    .catch(showError);

                            }}>Done
                            </button>
                        )}
                        {assignment.type == "task" && assignment.status == "done" && (
                            <button onClick={() => {
                                shpanKidsApi.updateTaskStatus({
                                    apiUpdateTaskStatusCommandArgs: {
                                        taskId: assignment.id,
                                        status: "open",
                                        forDate: assignment.forDate
                                    }
                                })
                                    .then(() => {
                                        setAssignments(assignments.map((t) => {
                                            if (t.id === assignment.id) {
                                                return {...t, status: "open"};
                                            }
                                            return t;
                                        }));
                                    })
                                    .catch(showError);

                            }}>Undo
                            </button>
                        )}
                        {assignment.type == "problemSet" && assignment.status == "open" && (
                            <button onClick={() => {
                                shpanKidsApi.loadProblemForAssignment({
                                    apiLoadProblemForAssignmentCommandArgs: {
                                        forDate: assignment.forDate,
                                        assignmentId: assignment.id
                                    }
                                })
                                    .then((p) => {
                                        props.uiCtx.showModal((
                                            <ProblemComponent
                                                uiCtx={props.uiCtx}
                                                problem={p.problem}
                                                submitAnswer={(answerId: string) => {
                                                    shpanKidsApi.submitProblemAnswer({
                                                        apiSubmitProblemAnswerCommandArgs: {
                                                            assignmentId: assignment.id,
                                                            answerId: answerId,
                                                            problemId: p.problem.id,
                                                        }
                                                    })
                                                        .then((res) => {
                                                            if (res.isCorrect) {
                                                                alert("That is Correct!");
                                                            } else {
                                                                const correctRes = p.problem.answers.find((a) => a.id === res.correctAnswerId);
                                                                if (correctRes) {
                                                                    let msg = "That is not correct. The correct answer is: " + correctRes.title
                                                                    if (res.explanation) {
                                                                        msg += "\n" + res.explanation!;
                                                                    }
                                                                    alert(msg);
                                                                } else {
                                                                    alert("That is not correct, not sure what is the correct answer");
                                                                }


                                                            }
                                                        })
                                                    .catch(showError)

                                                    props.uiCtx.hideModal();
                                                }}
                                            />))
                                    })
                                    .catch(showError);

                            }}>Load Problem
                            </button>
                        )}
                    </li>

                ))}
            </ol>

        </div>
    );

}
export default AssignmentsPage;