import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiTaskStats, GetStatsRequest} from "../../openapi";


export interface StatsPageProps {
    uiCtx: UiCtx;
}

enum taskPeriod {
    DAY = "day",
    WEEK = "week",
    MONTH = "month",
}
function formatDate(date: Date): string {
    const day = String(date.getDate()).padStart(2, "0");
    const month = String(date.getMonth() + 1).padStart(2, "0");
    return `${day}/${month}`;
}


const StatsPage: React.FC<StatsPageProps> = (props) => {

    const [taskStats, setTaskStats] = React.useState<ApiTaskStats[]>();
    const [loading, setLoading] = React.useState(true);
    const [period, setPeriod] = React.useState<taskPeriod>(taskPeriod.DAY);


    React.useEffect(() => {


        let requestParameters: GetStatsRequest | undefined
        switch (period) {
            case taskPeriod.WEEK:
                requestParameters = {
                    from: new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000),
                    to: new Date(),
                };
                break;
            case taskPeriod.MONTH:
                requestParameters = {
                    from: new Date(new Date().getTime() - 30 * 24 * 60 * 60 * 1000),
                    to: new Date(),
                };
                break;
        }

        shpanKidsApi.getStats(requestParameters)
            .then(setTaskStats)
            .then(() => setLoading(false))
            .catch(showError);

    }, [period]);

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
        .sort((a, b) => b - a)

    // create map of family members by email
    return (
        <div>
            <h2>Task Statistics</h2>
            <div>
                <label>Period:</label>
                <select value={period} onChange={(e) => setPeriod(e.target.value as taskPeriod)}>
                    <option value={taskPeriod.DAY}>Day</option>
                    <option value={taskPeriod.WEEK}>Week</option>
                    <option value={taskPeriod.MONTH}>Month</option>
                </select>
            </div>
            {loading && <div>Loading...</div>}
            <div>
                <h3>Stats</h3>
                <table>
                    <thead>
                    <tr>
                        <th>User</th>
                        {sortedDates.map((dateNum) => (
                            <th key={dateNum.toString()}>{formatDate(new Date(dateNum))}</th>
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