import { lazy } from 'react';
import { Route, Routes } from 'react-router-dom';

const Swap = lazy(() => import('@/app/views/Swap'));
const TransactionsHistory = lazy(() => import('@app/views/TransactionsHistory'));
const NotFound = lazy(() => import('@app/views/NotFound'));

/**
 * ComponentRoutes describes location mapping with components.
 */
export class ComponentRoutes {
    constructor(
        public path: string,
        public component: React.ReactNode,
        public children?: ComponentRoutes[]
    ) { }

    /** with is method that creates child sub routes path */
    public with(
        child: ComponentRoutes,
        parrent: ComponentRoutes
    ): ComponentRoutes {
        child.path = `${parrent.path}/${child.path}`;

        return this;
    }

    /** addChildren is method that adds children components to component */
    public addChildren(children: ComponentRoutes[]): ComponentRoutes {
        this.children = children.map((child: ComponentRoutes) =>
            child.with(child, this)
        );

        return this;
    }
}

/**
 * RoutesConfig contains information about all routes and subroutes.
 */
export class RoutesConfig {
    public static Swap: ComponentRoutes = new ComponentRoutes(
        '/',
        <Swap/>,
    );
    static TransactionsHistory: ComponentRoutes = new ComponentRoutes(
        '/history',
        <TransactionsHistory/>,
    );
    public static NotFound: ComponentRoutes = new ComponentRoutes(
        '/*',
        <NotFound/>,
    );

    /** Routes is an array of logical router components */
    public static routes: ComponentRoutes[] = [
        RoutesConfig.Swap,
        RoutesConfig.TransactionsHistory,
        RoutesConfig.NotFound,
    ];
}

export const RootRoutes = () =>
    <Routes>
        {RoutesConfig.routes.map(
            (route: ComponentRoutes, index: number) =>
                <Route
                    key={index}
                    path={route.path}
                    element={route.component}
                />
        )}
    </Routes>;
