import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiAssignment} from "../../openapi";


export interface TasksPageProps {
    uiCtx: UiCtx;
}

const AssignmentsPage: React.FC<TasksPageProps> = (props) => {
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
                        {assignment.type == "task" &&  assignment.status == "open" && (
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
                    </li>

                ))}
            </ol>

        </div>
    );

}
export default AssignmentsPage;