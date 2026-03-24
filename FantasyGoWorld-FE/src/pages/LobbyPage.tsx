import React, { useEffect, useState } from 'react';
import { useAuthStore } from '../store/authStore';
import { useLobbyStore } from '../store/lobbyStore';
import { logout } from '../api/auth';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

export const LobbyPage: React.FC = () => {
    const user = useAuthStore((s) => s.currentUser);
    const clearAuth = useAuthStore((s) => s.clearAuth);
    const { onlineUsers, publicRooms, totalUsers, totalRooms, loading, fetchLobbyData, createNewRoom } = useLobbyStore();
    const navigate = useNavigate();

    const [isCreating, setIsCreating] = useState(false);
    const [roomDesc, setRoomDesc] = useState('');
    const [isPublic, setIsPublic] = useState(true);
    const [password, setPassword] = useState('');

    useEffect(() => {
        fetchLobbyData();
        // 简单轮询刷新以模拟大厅动态，正式环境(P4)会使用 WebSocket 推送
        const interval = setInterval(() => {
            fetchLobbyData();
        }, 10000);
        return () => clearInterval(interval);
    }, [fetchLobbyData]);

    const handleLogout = async () => {
        try {
            await logout();
        } catch (e: any) {
            console.error(e);
        } finally {
            clearAuth();
            navigate('/login');
            toast.success('已退出登录');
        }
    };

    const handleCreateRoom = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await createNewRoom({ description: roomDesc, is_public: isPublic, password });
            toast.success('房间创建成功！');
            setIsCreating(false);
            setRoomDesc('');
            setPassword('');
        } catch (e: any) {
            toast.error(e.message || '创建房间失败');
        }
    };

    return (
        <div style={{ padding: '20px', maxWidth: '1200px', margin: '0 auto', color: 'white', minHeight: '100vh', boxSizing: 'border-box' }}>
            <header style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '40px', borderBottom: '1px solid rgba(255,255,255,0.1)', paddingBottom: '20px' }}>
                <h1>棋幻大厅</h1>
                {user && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: '15px' }}>
                        <span>欢迎, <strong>{user.nickname}</strong> ({user.rank})</span>
                        <button onClick={handleLogout} style={{ padding: '6px 12px', background: 'rgba(255, 75, 43, 0.8)', border: 'none', color: 'white', borderRadius: '4px', cursor: 'pointer' }}>退出登录</button>
                    </div>
                )}
            </header>

            <div style={{ display: 'grid', gridTemplateColumns: 'minmax(300px, 1fr) 300px', gap: '30px' }}>
                <section>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <h2>对弈房间 ({totalRooms})</h2>
                        <button onClick={() => setIsCreating(!isCreating)} style={{ padding: '8px 16px', background: '#4facfe', border: 'none', borderRadius: '8px', color: 'white', cursor: 'pointer', fontWeight: 'bold' }}>
                            {isCreating ? '取消创建' : '+ 创建新对局'}
                        </button>
                    </div>

                    {isCreating && (
                        <form onSubmit={handleCreateRoom} style={{ background: 'rgba(255,255,255,0.05)', padding: '20px', borderRadius: '12px', marginTop: '20px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
                            <input value={roomDesc} onChange={(e) => setRoomDesc(e.target.value)} placeholder="房间描述 (如：新手来战！)" style={{ padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.2)', background: 'transparent', color: 'white' }} required />
                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                                <label>
                                    <input type="checkbox" checked={!isPublic} onChange={(e) => setIsPublic(!e.target.checked)} /> 设为私密房间
                                </label>
                            </div>
                            {!isPublic && <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="设置密码" style={{ padding: '10px', borderRadius: '8px', border: '1px solid rgba(255,255,255,0.2)', background: 'transparent', color: 'white' }} required />}
                            <button type="submit" style={{ padding: '10px', background: '#005bea', border: 'none', borderRadius: '8px', color: 'white', cursor: 'pointer', marginTop: '10px' }}>确认创建</button>
                        </form>
                    )}

                    <div style={{ marginTop: '20px', display: 'flex', flexDirection: 'column', gap: '15px' }}>
                        {loading && publicRooms.length === 0 ? <p>加载中...</p> : null}
                        {!loading && publicRooms.length === 0 ? <p style={{ color: '#aaa' }}>暂时没有人在下棋哦，快去创建一个吧！</p> : null}
                        {publicRooms.map(room => (
                            <div key={room.id} style={{ background: 'rgba(255,255,255,0.08)', padding: '20px', borderRadius: '12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                <div>
                                    <h3 style={{ margin: '0 0 8px 0' }}>{room.description || '无描述'}</h3>
                                    <p style={{ margin: 0, color: '#bbb', fontSize: '0.9em' }}>房主: {room.owner?.nickname} ({room.owner?.rank}) · 状态: {room.status === 0 ? '等待中' : '对局中'}</p>
                                </div>
                                <button style={{ padding: '8px 24px', background: 'white', color: 'black', border: 'none', borderRadius: '20px', cursor: 'pointer', fontWeight: 'bold' }}>加入</button>
                            </div>
                        ))}
                    </div>
                </section>

                <section>
                    <h2>在线棋手 ({totalUsers})</h2>
                    <div style={{ background: 'rgba(0,0,0,0.2)', borderRadius: '12px', padding: '20px', height: '600px', overflowY: 'auto' }}>
                        {onlineUsers.map(u => (
                            <div key={u.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid rgba(255,255,255,0.05)' }}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                                    <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#4facfe', boxShadow: '0 0 8px #4facfe' }}></div>
                                    <span>{u.nickname}</span>
                                </div>
                                <span style={{ color: '#888', fontSize: '0.9em' }}>{u.rank}</span>
                            </div>
                        ))}
                    </div>
                </section>
            </div>
        </div>
    );
};
