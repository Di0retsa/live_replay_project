import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/Login';
import HomePage from './components/HomePage';
import UploadVideo from './components/UploadVideo';
import PlayVideo from './components/PlayVideo';
import './App.css';
import './utils/api';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState('');
  const [isLoadingAuth, setIsLoadingAuth] = useState(true); // 新增：认证加载状态

  useEffect(() => {
    const token = localStorage.getItem('token');
    const savedUsername = localStorage.getItem('username');
    if (token && savedUsername) {
      setIsLoggedIn(true);
      setUsername(savedUsername);
    }
    setIsLoadingAuth(false); // 无论如何，认证检查完成
  }, []);

  const handleLoginSuccess = (userData) => {
    setIsLoggedIn(true);
    setUsername(userData.username || userData.phone);
    localStorage.setItem('token', userData.token);
    localStorage.setItem('username', userData.username || userData.phone);
  };

  const handleLogout = () => {
    setIsLoggedIn(false);
    setUsername('');
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    // 可以选择在这里导航到登录页
    navigate('/login');
  };

  // 如果仍在检查认证状态，显示加载提示或 null
  if (isLoadingAuth) {
    return <div>Loading authentication...</div>; // 或者返回 null，或者一个骨架屏
  }

  return (
    <Router>
      <Routes>
        <Route
          path="/login"
          element={!isLoggedIn ? <Login onLoginSuccess={handleLoginSuccess} /> : <Navigate to="/" replace />}
        />
        <Route
          path="/"
          element={isLoggedIn ? <HomePage username={username} onLogout={handleLogout} /> : <Navigate to="/login" replace />}
        />
        <Route
          path="/upload"
          element={isLoggedIn ? <UploadVideo /> : <Navigate to="/login" replace />}
        />
        <Route
          path="/replay/:replayId"
          element={isLoggedIn ? <PlayVideo /> : <Navigate to="/login" replace />}
        />
        {/* 你可能还需要一个404页面或者一个默认重定向到首页（如果已登录）或登录页（如果未登录） */}
        {/* 例如: <Route path="*" element={<Navigate to={isLoggedIn ? "/" : "/login"} replace />} /> */}
      </Routes>
    </Router>
  );
}

export default App;
