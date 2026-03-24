import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { login, register } from '../api/auth';
import { useAuthStore } from '../store/authStore';
import toast, { Toaster } from 'react-hot-toast';

export const LoginPage: React.FC = () => {
    const [isLogin, setIsLogin] = useState(true);
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [nickname, setNickname] = useState('');
    const [rank, setRank] = useState('18K');
    const [loading, setLoading] = useState(false);

    const setAuth = useAuthStore((s) => s.setAuth);
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);

        try {
            if (isLogin) {
                const res = await login({ username, password });
                setAuth(res.data.access_token, { ...res.data.user, dnd: false });
                toast.success(`欢迎回来，${res.data.user.nickname}`);
                navigate('/lobby');
            } else {
                await register({ username, password, nickname, rank });
                toast.success('注册成功！快去登录吧~');
                setIsLogin(true);
            }
        } catch (err: any) {
            toast.error(err.message || '操作失败，请稍后重试');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="login-container">
            <Toaster position="top-right" />
            <div className="glass-card">
                <h1 className="title">Fantasy Go World</h1>
                <p className="subtitle">{isLogin ? '开启你的棋幻对决' : '加入棋幻世界'}</p>

                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="input-group">
                        <label>用户名</label>
                        <input
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            required
                            placeholder="请输入用户名"
                        />
                    </div>

                    <div className="input-group">
                        <label>密码</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                            placeholder="请输入密码"
                        />
                    </div>

                    {!isLogin && (
                        <>
                            <div className="input-group">
                                <label>昵称</label>
                                <input
                                    type="text"
                                    value={nickname}
                                    onChange={(e) => setNickname(e.target.value)}
                                    required
                                    placeholder="请输入昵称"
                                />
                            </div>
                            <div className="input-group">
                                <label>棋力 (18K-9D)</label>
                                <select value={rank} onChange={(e) => setRank(e.target.value)}>
                                    <option value="18K">18K</option>
                                    <option value="10K">10K</option>
                                    <option value="5K">5K</option>
                                    <option value="1K">1K</option>
                                    <option value="1D">1D</option>
                                    <option value="5D">5D</option>
                                    <option value="9D">9D</option>
                                </select>
                            </div>
                        </>
                    )}

                    <button type="submit" className="submit-btn" disabled={loading}>
                        {loading ? '处理中...' : (isLogin ? '登录' : '注册')}
                    </button>
                </form>

                <div className="toggle-auth">
                    <button onClick={() => setIsLogin(!isLogin)}>
                        {isLogin ? '还没有账号？去注册' : '已有账号？去登录'}
                    </button>
                </div>
            </div>

            <style>{`
        .login-container {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
          font-family: 'Inter', system-ui, sans-serif;
          color: white;
          padding: 20px;
        }
        .glass-card {
          background: rgba(255, 255, 255, 0.05);
          backdrop-filter: blur(10px);
          border: 1px solid rgba(255, 255, 255, 0.1);
          border-radius: 20px;
          padding: 40px;
          width: 100%;
          max-width: 400px;
          box-sizing: border-box;
          box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }
        .title {
          font-size: 2.5rem;
          font-weight: 800;
          text-align: center;
          margin-bottom: 8px;
          background: linear-gradient(to right, #4facfe 0%, #00f2fe 100%);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }
        .subtitle {
          text-align: center;
          color: rgba(255, 255, 255, 0.6);
          margin-bottom: 32px;
        }
        .auth-form {
          display: flex;
          flex-direction: column;
          gap: 20px;
        }
        .input-group {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }
        .input-group label {
          font-size: 0.9rem;
          font-weight: 500;
          color: rgba(255, 255, 255, 0.8);
        }
        .input-group input, .input-group select {
          padding: 12px 16px;
          border-radius: 12px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          background: rgba(255, 255, 255, 0.05);
          color: white;
          font-size: 1rem;
          transition: all 0.2s;
        }
        .input-group input:focus, .input-group select:focus {
          outline: none;
          border-color: #4facfe;
          background: rgba(255, 255, 255, 0.1);
        }
        .submit-btn {
          margin-top: 10px;
          padding: 14px;
          border-radius: 12px;
          border: none;
          background: linear-gradient(to right, #00c6fb 0%, #005bea 100%);
          color: white;
          font-size: 1.1rem;
          font-weight: 600;
          cursor: pointer;
          transition: all 0.3s;
        }
        .submit-btn:hover {
          transform: translateY(-2px);
          box-shadow: 0 10px 20px -10px #005bea;
        }
        .submit-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }
        .toggle-auth {
          margin-top: 24px;
          text-align: center;
        }
        .toggle-auth button {
          background: none;
          border: none;
          color: #4facfe;
          cursor: pointer;
          font-size: 0.9rem;
          transition: color 0.2s;
        }
        .toggle-auth button:hover {
          color: #00f2fe;
          text-decoration: underline;
        }
      `}</style>
        </div>
    );
};
