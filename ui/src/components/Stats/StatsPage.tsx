import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiTaskStats} from "../../openapi";


export interface StatsPageProps {
    uiCtx: UiCtx;
}

const StatsPage: React.FC<StatsPageProps> = (props) => {

    const [taskStats, setTaskStats] = React.useState<ApiTaskStats[]>();
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        shpanKidsApi.getStats()
            .then(setTaskStats)
            .then(() => setLoading(false))
            .catch(showError);

    }, []);

    // tasksByUserId will have by user sorted tasks
    const tasksByUserId = new Map<string, Map<number, ApiTaskStats>>();
    const allDates = new Set<number>();
    if (taskStats) {
        taskStats!.forEach((task) => {
            allDates.add(task.forDate.setHours(0, 0, 0, 0));
            if (!tasksByUserId.has(task.userId)) {
                tasksByUserId.set(task.userId, new Map());
            }
            tasksByUserId.get(task.userId)!.set(task.forDate.setHours(0, 0, 0, 0), task);
        });
    }

    const sortedDates = Array.from(allDates)
        .sort((a, b) => a - b)

    // create map of family members by email
    return (
        <div>
            <h2>Task Statistics</h2>
            {loading && <div>Loading...</div>}
            <div>
                <h3>Stats</h3>
                <table>
                    <thead>
                    <tr>
                        <th>User</th>
                        {sortedDates.map((dateNum) => (
                            <th key={dateNum.toString()}>{new Date(dateNum).toLocaleDateString("en-GB")}</th>
                        ))}
                    </tr>
                    </thead>
                    <tbody>
                    {Array.from(tasksByUserId.entries()).map((te) => (
                        <tr key={te[0]}>
                            <td>{te[0]}</td>
                            {sortedDates.map((dateNum) => {
                                const stats = te[1].get(dateNum);
                                return stats ? (
                                        <td key={dateNum}
                                            style={{backgroundColor: stats!.doneTasksCount === stats!.totalTasksCount ? 'green' : 'red'}}>
                                            {stats!.doneTasksCount}/{stats!.totalTasksCount}
                                        </td>
                                    ) :
                                    (<td key={dateNum}>N/A</td>);
                            })}
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>
        </div>
    );

}
export default StatsPage;