import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiTask} from "../../openapi";


export interface TasksPageProps {
    uiCtx: UiCtx;
}

const TasksPage: React.FC<TasksPageProps> = (props) => {
    const [tasks, setTasks] = React.useState<ApiTask[]>([]);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        shpanKidsApi.listTasks()
            .then(tasks => tasks.sort((a, b) => a.id.localeCompare(b.id)))
            .then(setTasks).then(() =>
            setLoading(false)
        )
            .catch(showError);

    }, [props.uiCtx]);

    return (
        <div>
            <h2>Today's tasks</h2>
            {loading && <div>Loading...</div>}
            <ol>
                {tasks.map((task) => (
                    <li style={{textDecoration: task.status == "done" ? "line-through" : "auto"}}
                        key={task.id}>{task.title}
                        {task.status == "open" && (
                            <button onClick={() => {
                                shpanKidsApi.updateTaskStatus({
                                    apiUpdateTaskStatusCommandArgs: {
                                        taskId: task.id,
                                        status: "done",
                                        forDate: task.forDate
                                    }
                                })
                                    .then(() => {
                                        setTasks(tasks.map((t) => {
                                            if (t.id === task.id) {
                                                return {...t, status: "done"};
                                            }
                                            return t;
                                        }));
                                    })
                                    .catch(showError);

                            }}>Done
                            </button>
                        )}
                        {task.status == "done" && (
                            <button onClick={() => {
                                shpanKidsApi.updateTaskStatus({
                                    apiUpdateTaskStatusCommandArgs: {
                                        taskId: task.id,
                                        status: "open",
                                        forDate: task.forDate
                                    }
                                })
                                    .then(() => {
                                        setTasks(tasks.map((t) => {
                                            if (t.id === task.id) {
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
export default TasksPage;