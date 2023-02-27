import { Suspense } from 'react';
import { BrowserRouter } from 'react-router-dom';

import { ThemeProvider } from '@app/components/common/ThemeProvider';
import { Navbar } from '@app/components/common/Navbar';
import { Notification } from '@app/components/common/Notification';
import { RootRoutes as Routes } from '@app/routes';

import './index.scss';

function App() {
    return (
        <Suspense fallback={<div>Loading...</div>}>
            <ThemeProvider>
                <BrowserRouter basename="/">
                    <Notification />
                    <Navbar />
                    <Routes />
                </BrowserRouter>
            </ThemeProvider>
        </Suspense>
    );
};

export default App;
