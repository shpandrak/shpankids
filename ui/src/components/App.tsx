import "./App.css";
import React, {useEffect, useState} from "react";
import {Link, Outlet, Route, Routes} from "react-router-dom";
import Modal from "./Common/Modal.tsx";
import UiCtx from "./Common/UiCtx.ts";
import AssignmentsPage from "./Assignment/AssignmentsPage.tsx";
import {Configuration, ShpankidsApi, UIApi, UIFamilyInfo, UIUserInfo} from "../openapi";
import FamilyPage from "./Family/FamilyPage.tsx";
import StatsPage from "./Stats/StatsPage.tsx";
import ProblemSetsPage from "./Problem/ProblemSetsPage.tsx";

export const shpanKidsApi = new ShpankidsApi(new Configuration({
    basePath: ""
}));
export const uiApi = new UIApi(new Configuration({
    basePath: ""
}));


function App() {
    // const [count, setCount] = useState(0)
    const [modal, setModal] = useState<React.JSX.Element>();
    const [userInfo, setUserInfo] = useState<UIUserInfo>();
    const [familyInfo, setFamilyInfo] = useState<UIFamilyInfo>();
    const [uiCtx] = useState<UiCtx>(
        new UiCtx(setModal)
    );

    useEffect(() => {
        uiApi.getUserInfo().then(setUserInfo);
        uiApi.getFamilyInfo().then(setFamilyInfo);
    }, []);

    return userInfo && familyInfo && (
        <div>
            <h3>Welcome {userInfo.firstName ?? userInfo!.email}!</h3>

            {modal && (
                <Modal
                    isOpen={true}
                    onClose={() => setModal(undefined)}
                    children={modal!}
                />
            )}
            <Routes>
                <Route path="/ui" element={(

                    <div>
                        <nav>
                            <span>
                              <Link to="">Today</Link>
                            </span>
                            <span>&nbsp;&nbsp;|&nbsp;&nbsp;</span>
                            <span>
                            <Link to="page/stats">Stats</Link>
                            </span>
                            {userInfo.role === "familyAdmin" && (
                                <>
                                    <span>&nbsp;&nbsp;|&nbsp;&nbsp;</span>
                                    <span>
                                    <Link to="page/family">Family</Link>
                                    </span>
                                    <span>&nbsp;&nbsp;|&nbsp;&nbsp;</span>
                                    <span>
                                    <Link to="page/problemsets">Problem Sets</Link>
                                    </span>
                                </>
                            )}
                        </nav>
                        <hr/>

                        {/* An <Outlet> renders whatever child route is currently active,
                            so you can think about this <Outlet> as a placeholder for
                            the child routes we defined above.
                        */}
                        <Outlet/>
                    </div>
                )}>
                    <Route
                        index
                        element={
                            <AssignmentsPage uiCtx={uiCtx}/>
                        }
                    />
                    <Route
                        path="page/family"
                        element={<FamilyPage uiCtx={uiCtx}/>}
                    />
                    <Route
                        path="page/stats"
                        element={<StatsPage uiCtx={uiCtx}></StatsPage>}
                    />
                    <Route
                        path="page/problemsets"
                        element={<ProblemSetsPage uiCtx={uiCtx} familyInfo={familyInfo}/>}
                    />
                </Route>
            </Routes>
        </div>
    );
}

function NoMatch() {
    return (
        <div>
            <h2>Nothing to see here!</h2>
            <p>
                <Link to="">Go to the home page</Link>
            </p>
        </div>
    );
}

export default App;
