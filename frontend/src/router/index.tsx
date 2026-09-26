import { Navigate, createBrowserRouter } from 'react-router-dom';
import App from '../App';
import TransportContainerPage from '../pages/TransportContainerPage';
import TemperatureWindowPage from '../pages/TemperatureWindowPage';
import ExcursionEventPage from '../pages/ExcursionEventPage';
import DispositionDecisionPage from '../pages/DispositionDecisionPage';
import AuditPage from '../pages/AuditPage';
import { getSession } from '../api/client';
import { roleAtLeast } from '../types/domain';
function AuditGuard() { return roleAtLeast(getSession()?.role, 'reviewer') ? <AuditPage /> : <Navigate to="/containers" replace />; }
export const router = createBrowserRouter([{ path: '/', element: <App />, children: [
  { index: true, element: <Navigate to="/containers" replace /> },
  { path: 'containers', element: <TransportContainerPage /> }, { path: 'windows', element: <TemperatureWindowPage /> }, { path: 'excursions', element: <ExcursionEventPage /> }, { path: 'dispositions', element: <DispositionDecisionPage /> },
  { path: 'audit', element: <AuditGuard /> },
] }], { future: { v7_fetcherPersist: true, v7_normalizeFormMethod: true, v7_partialHydration: true, v7_relativeSplatPath: true, v7_skipActionErrorRevalidation: true } });
