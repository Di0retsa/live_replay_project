import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import viewLogo from '/dianshizhiboguankanshipinbofangmeiti.svg'
import commentLogo from '/31pinglun.svg'
import '../styles/HomePage.css';
import api from '../utils/api';

function HomePage({ username, onLogout }) {
  const navigate = useNavigate();
  const [replays, setReplays] = useState([]);
  const [loading, setLoading] = useState(true);
  const [token] = useState(localStorage.getItem('token') || null);
  const [error, setError] = useState('');
  const [expandedVideo, setExpandedVideo] = useState(null);
  const [expandedDescriptions, setExpandedDescriptions] = useState({});

  useEffect(() => {
    // 获取所有视频数据
    api.get('/replay/list')
      .then(response => {
        // 正确获取 List 数组
        const replayArray = Array.isArray(response.data.data?.List) ? response.data.data.List : [];
        setReplays(replayArray);
        setLoading(false);
      })
      .catch(error => {
        console.error('获取视频列表失败:', error);
        setError(error.message);
        setLoading(false);
      });
  }, [token]);

  const handleUploadVideo = () => {
    navigate('/upload');
  };

  const handleVideoClick = (videoId) => {
    if (expandedVideo === videoId) {
      setExpandedVideo(null);
    } else {
      setExpandedVideo(videoId);
    }
  };

  const formatDuration = (seconds) => {
    if (!seconds || isNaN(seconds)) return '未知';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remainingSeconds = Math.floor(seconds % 60);

    const pad = (num) => num.toString().padStart(2, '0');

    if (hours > 0) {
      return `${hours}:${pad(minutes)}:${pad(remainingSeconds)}`;
    } else {
      return `${pad(minutes)}:${pad(remainingSeconds)}`;
    }
  };

  const formatDateTime = (timestamp) => {
    if (!timestamp) return '未知';

    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return '未知';

    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false
    });
  };

  const toggleDescription = (event, videoId) => {
    event.stopPropagation();
    setExpandedDescriptions(prev => ({
      ...prev,
      [videoId]: !prev[videoId]
    }));
  };

  const formatDescription = (description, videoId) => {
    if (!description) return '暂无简介';

    const isExpanded = expandedDescriptions[videoId];
    const maxLength = 50;

    if (description.length <= maxLength || isExpanded) {
      return description;
    }

    return (
      <>
        {description.substring(0, maxLength)}...
        <button
          className="show-more-btn"
          onClick={(e) => toggleDescription(e, videoId)}
        >
          显示更多
        </button>
      </>
    );
  };

  const handlePlayVideo = (e, replay) => {
    e.stopPropagation();
    navigate(`/replay/${replay.replay_id}`, { state: { replay } });
  };

  return (
    <div className="home-container">
      {/* 固定顶部栏 */}
      <header className="header">
        <div className="user-info">
          <span>欢迎，{username || '用户'}</span>
        </div>
        <div className="header-actions">
          <button onClick={onLogout} className="logout-btn">退出登录</button>
        </div>
      </header>

      {/* 视频内容区域 */}
      <main className="content">
        {loading ? (
          <div className="loading">加载中...</div>
        ) : error ? (
          <div className="error-message">{error}</div>
        ) : (
          <div className="video-grid">
            {replays.map(replay => (
              <div
                key={replay.replay_id}
                className={`video-card ${expandedVideo === replay.replay_id ? 'expanded' : ''}`}
                onClick={() => handleVideoClick(replay.replay_id)}
              >
                <div className="video-thumbnail">
                  <img src={`http://localhost:8080/replay/${replay.cover_path}`} alt={replay.title} />
                </div>
                <div className="video-info">
                  <div className='info'>
                    <div className='info-left'>
                      <h3>{replay.title}</h3>
                      <p>
                        <img src={viewLogo} className="icon" alt="viewIcon" />
                        <span> </span>
                        {replay.views}
                        <span> </span>
                        <img src={commentLogo} className="icon" alt="commentIcon" />
                        <span> </span>
                        {replay.comments}
                      </p>
                    </div>
                    {expandedVideo === replay.replay_id && (
                      <div className='info-right'>
                        <button className="play-btn" onClick={(e) => handlePlayVideo(e, replay)}>
                          <span className="play-icon">▶</span>
                        </button>
                      </div>
                    )}
                  </div>
                  {expandedVideo === replay.replay_id && (
                    <div className="video-details">
                      <p className="video-description">
                        {formatDescription(replay.description, replay.replay_id)}
                        {expandedDescriptions[replay.replay_id] && (
                          <button
                            className="show-less-btn"
                            onClick={(e) => toggleDescription(e, replay.replay_id)}
                          >
                            收起
                          </button>
                        )}
                      </p>
                      <div className="video-meta">
                        <span className="video-duration">时长: {formatDuration(replay.duration)}</span>
                        <span className="video-date">发布时间: {formatDateTime(replay.create_time)}</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* 固定底部栏 */}
      <footer className="footer">
        <button className="upload-btn" onClick={handleUploadVideo}>
          <span className="upload-icon">+</span>
        </button>
      </footer>
    </div>
  );
}

export default HomePage;