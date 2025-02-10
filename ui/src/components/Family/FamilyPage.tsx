import * as React from "react";
import UiCtx from "../Common/UiCtx.ts";
import {shpanKidsApi, uiApi} from "../App.tsx";
import {showError} from "../Util.ts";
import {ApiTask, UIFamilyInfo} from "../../openapi";


export interface FamilyPageProps {
    uiCtx: UiCtx;
}

const FamilyPage: React.FC<FamilyPageProps> = (props) => {

    const [familyInfo, setFamilyInfo] = React.useState<UIFamilyInfo>();
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        uiApi.getFamilyInfo()
            .then(setFamilyInfo).then(() =>
            setLoading(false)
        )
            .catch(showError);

    }, []);

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

            </div>}
        </div>
    );

}
export default FamilyPage;