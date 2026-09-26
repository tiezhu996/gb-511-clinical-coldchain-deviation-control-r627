import { FormEvent, useState } from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import DeviceThermostatIcon from '@mui/icons-material/DeviceThermostat';

type LoginProps = { loading: boolean; onLogin: (username: string, password: string) => Promise<unknown> };

export default function LoginPage({ loading, onLogin }: LoginProps) {
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('Admin123!');
  const [error, setError] = useState('');
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    setError('');
    try { await onLogin(username, password); } catch (reason) { setError(reason instanceof Error ? reason.message : String(reason)); }
  };
  return <main className="login-shell">
    <section className="login-context"><div className="login-mark"><DeviceThermostatIcon /></div><p>COLD CHAIN QUALITY</p><h1>临床冷链<br />温度偏差处置</h1><span>传感器证据、质量复核与处置决定保持同一条审计链。</span><div className="login-signals"><strong>PostgreSQL</strong><strong>Redis</strong><strong>MinIO</strong></div></section>
    <section className="login-panel"><form onSubmit={submit}><LockOutlinedIcon className="login-lock" /><p>QUALITY CONTROL DESK</p><h2>登录质量工作台</h2><TextField label="用户名" value={username} onChange={(event) => setUsername(event.target.value)} fullWidth /><TextField label="密码" type="password" value={password} onChange={(event) => setPassword(event.target.value)} fullWidth />{error && <div className="alert" role="alert">{error}</div>}<Button type="submit" variant="contained" disabled={loading} fullWidth>{loading ? '正在验证…' : '进入工作台'}</Button><small>演示账号：admin / reviewer / operator / viewer，密码均为 Admin123!</small></form></section>
  </main>;
}
