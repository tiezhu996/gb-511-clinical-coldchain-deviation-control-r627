
import { NavLink, Outlet } from 'react-router-dom';
import AcUnitIcon from '@mui/icons-material/AcUnit';
import Inventory2OutlinedIcon from '@mui/icons-material/Inventory2Outlined';
import RuleOutlinedIcon from '@mui/icons-material/RuleOutlined';
import WarningAmberOutlinedIcon from '@mui/icons-material/WarningAmberOutlined';
import GavelOutlinedIcon from '@mui/icons-material/GavelOutlined';
import HistoryOutlinedIcon from '@mui/icons-material/HistoryOutlined';
import { useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import { roleAtLeast } from './types/domain';
const navigation = [
  { to: '/containers', label: '运输容器', icon: Inventory2OutlinedIcon },
  { to: '/windows', label: '温控规则', icon: RuleOutlinedIcon },
  { to: '/excursions', label: '偏差事件', icon: WarningAmberOutlinedIcon },
  { to: '/dispositions', label: '处置决定', icon: GavelOutlinedIcon },
  { to: '/audit', label: '审计记录', icon: HistoryOutlinedIcon, minimum: 'reviewer' as const },
];
export default function App() {
  const { session, loading, signIn, logout } = useAuth();
  if (!session) return <LoginPage loading={loading} onLogin={signIn} />;
  const items = navigation.filter((item) => !item.minimum || roleAtLeast(session.role, item.minimum));
  return <div className="app-shell"><aside><div className="brand"><AcUnitIcon /><span>QUALITY CONTROL</span><strong>临床冷链偏差处置</strong></div><nav>{items.map((item) => { const Icon = item.icon; return <NavLink key={item.to} to={item.to}><Icon />{item.label}</NavLink>; })}</nav><div className="user-panel"><span>{session.displayName}</span><small>{session.role}</small><button onClick={logout}>退出登录</button></div></aside><section className="content"><header className="topbar"><span>临床配送质量中心</span><span className="live-dot">传感器链路在线</span></header><Outlet /></section></div>;
}
