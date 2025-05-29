import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import '../styles/PlayVideo.css';
import api from '../utils/api';
import Picker from 'emoji-picker-react';

function PlayVideo() {
  const { replayId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const [replay, setReplay] = useState(location.state?.replay || null);
  const [token] = useState(localStorage.getItem('token') || null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showOverlay, setShowOverlay] = useState(true);
  const videoRef = useRef(null);

  const [isDragging, setIsDragging] = useState(false);
  const [startX, setStartX] = useState(0);
  const [startTime, setStartTime] = useState(0);
  const [dragDirection, setDragDirection] = useState(null); // 'left' 或 'right'
  const [dragDistance, setDragDistance] = useState(0);
  const [showSeekIndicator, setShowSeekIndicator] = useState(false);
  const [seekTime, setSeekTime] = useState(0);

  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [currentUser, setCurrentUser] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const ws = useRef(null);
  const messagesEndRef = useRef(null);
  const typingAreaRef = useRef(null);

  const [isPlaying, setIsPlaying] = useState(true);

  useEffect(() => {
    if (!token) {
      alert('请先登录');
      navigate('/login');
    }
    setCurrentUser(localStorage.getItem('username') || '我');
  }, [token, navigate]);

  useEffect(() => {
    if (!token || !replayId) return;

    setLoading(true);
    console.log('正在获取视频信息...replayId:', replayId);

    api.get(`/replay/${replayId}`)
      .then(response => {
        console.log('API返回数据:', response.data);
        if (response.data.code === 200 && response.data.data) {
          setReplay(response.data.data);
          setError('');
        } else {
          throw new Error(response.data.msg || '获取视频信息失败');
        }
      })
      .catch(error => {
        console.error('获取视频信息失败:', error);
        setError(error.message);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [replayId, token]);

  useEffect(() => {
    const videoElement = videoRef.current;
    if (!videoElement) return;

    const handlePlay = () => setIsPlaying(true);
    const handlePause = () => setIsPlaying(false);
    const handleEnded = () => setIsPlaying(false);

    videoElement.addEventListener('play', handlePlay);
    videoElement.addEventListener('pause', handlePause);
    videoElement.addEventListener('ended', handleEnded);

    return () => {
      videoElement.removeEventListener('play', handlePlay);
      videoElement.removeEventListener('pause', handlePause);
      videoElement.removeEventListener('ended', handleEnded);
    };
  }, [videoRef.current]);

  const handlePlayClick = () => {
    setShowOverlay(false);
    if (videoRef.current) {
      videoRef.current.play();
      setIsPlaying(true);
    }
  };

  const handleBackClick = () => {
    navigate('/');
  };

  const formatTime = (seconds) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const handleVideoClick = (e) => {
    if (!videoRef.current || showOverlay) return;

    e.stopPropagation();

    if (videoRef.current.paused || videoRef.current.ended) {
      videoRef.current.play();
    } else {
      videoRef.current.pause();
    }
  };

  const handleTouchStart = (e) => {
    if (!videoRef.current || showOverlay) return;

    setIsDragging(true);
    setStartX(e.touches[0].clientX);
    setStartTime(videoRef.current.currentTime);
    setDragDirection(null);
    setDragDistance(0);
  };

  const handleTouchMove = (e) => {
    if (!isDragging || !videoRef.current || showOverlay) return;

    const currentX = e.touches[0].clientX;
    const distance = currentX - startX;
    const direction = distance > 0 ? 'right' : 'left';

    setDragDirection(direction);
    setDragDistance(Math.abs(distance));

    const timeChange = (distance / 100) * 10;
    const newTime = Math.max(0, Math.min(videoRef.current.duration, startTime + timeChange));

    setSeekTime(newTime);
    setShowSeekIndicator(true);
  };

  const handleTouchEnd = (e) => {
    if (!videoRef.current || showOverlay) return;

    if (isDragging && dragDistance < 10) {
      handleVideoClick(e);
    }
    else if (isDragging && dragDirection && dragDistance > 20) {
      videoRef.current.currentTime = seekTime;
    }

    setIsDragging(false);
    setDragDirection(null);
    setDragDistance(0);
    setShowSeekIndicator(false);
  };

  const handleMouseDown = (e) => {
    if (!videoRef.current || showOverlay) return;

    setIsDragging(true);
    setStartX(e.clientX);
    setStartTime(videoRef.current.currentTime);
    setDragDirection(null);
    setDragDistance(0);
  };

  const handleMouseMove = (e) => {
    if (!isDragging || !videoRef.current || showOverlay) return;

    const currentX = e.clientX;
    const distance = currentX - startX;
    const direction = distance > 0 ? 'right' : 'left';

    setDragDirection(direction);
    setDragDistance(Math.abs(distance));

    const timeChange = (distance / 100) * 10;
    const newTime = Math.max(0, Math.min(videoRef.current.duration, startTime + timeChange));

    setSeekTime(newTime);
    setShowSeekIndicator(true);
  };

  const handleMouseUp = (e) => {
    if (!videoRef.current || showOverlay) {
      if (isDragging) setIsDragging(false);
      return;
    }

    const localIsDragging = isDragging;
    const localDragDistance = dragDistance;
    const localDragDirection = dragDirection;

    if (localIsDragging && localDragDistance < 10) {
      handleVideoClick(e);
    }
    else if (localIsDragging && localDragDirection && localDragDistance > 20) {
      videoRef.current.currentTime = seekTime;
    }

    setIsDragging(false);
    setDragDirection(null);
    setDragDistance(0);
    setShowSeekIndicator(false);
  };

  const scrollToBottom = () => {
    if (messagesEndRef.current) {
      setTimeout(() => {
        messagesEndRef.current.scrollTop = messagesEndRef.current.scrollHeight;
      }, 0);
    }
  };

  useEffect(scrollToBottom);

  useEffect(() => {
    if (!replayId || !token) {
      console.log('replayId or token is missing');
      setIsConnected(false);
      return;
    }

    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.close(1000, "ReplayId has changed");
    }

    const wsUrl = `ws://localhost:8080/ws/chat/${replayId}?auth=${token}`;
    console.log('Connecting to WebSocket:', wsUrl);
    ws.current = new WebSocket(wsUrl);

    setIsConnected(false);
    setError(null);

    ws.current.onopen = () => {
      console.log(`WebSocket connected to room :${replayId}`);
      setIsConnected(true);
      setError(null);
    };

    ws.current.onmessage = (event) => {
      try {
        console.log('Raw WebSocket data received:', event.data);
        const rawData = event.data;
        const potentialJsons = rawData.trim().split('}{');

        potentialJsons.forEach((jsonString, index) => {
          let parsableJsonString = jsonString;
          if (index > 0) {
            parsableJsonString = '{' + parsableJsonString;
          }
          if (index < potentialJsons.length - 1) {
            parsableJsonString = parsableJsonString + '}';
          }

          try {
            const receivedMessage = JSON.parse(parsableJsonString);
            console.log('Parsed individual message:', receivedMessage);
            setMessages((prevMessages) => [...prevMessages, {
              sender: receivedMessage.username || '未知用户',
              content: receivedMessage.content,
              type: receivedMessage.type,
            }]);
          } catch (parseError) {
            console.error('Error parsing individual JSON string:', parseError);
            console.error('Problematic individual string:', parsableJsonString);
          }
        });

      } catch (error) {
        console.error('Error processing WebSocket message (outer catch):', error);
      }
    };

    ws.current.onerror = (event) => {
      console.log('WebSocket error:', event.code, event.reason);
      setIsConnected(false);
      setError("出现未知错误，请稍后再试");
    };

    ws.current.onclose = (event) => {
      console.log(`WebSocket closed from room ${replayId}:`, event.code, event.reason);
      setIsConnected(false);
      if (event.code !== 1000 && event.code !== 1005) {
        console.log('WebSocket closed normally');
        setError("连接已断开，请尝试刷新");
        return;
      } else {
        setError(null);
      }
    };
    return () => {
      if (ws.current) {
        console.log('Closing WebSocket connection');
        ws.current.close(1000, 'Closing WebSocket connection');
      }
    };
  }, [replayId, token]);

  const handleSendMessage = useCallback(() => {
    if (!isConnected) {
      alert('聊天未连接，无法发送消息！');
      return;
    }
    if (inputMessage.trim() === '') return;

    const messagePayload = {
      replay_id: parseInt(replayId),
      content: inputMessage.trim(),
      type: 'user-message',
    };

    setMessages((prevMessages) => [...prevMessages, {
      sender: currentUser,
      content: inputMessage.trim(),
      type: 'user-message',
    }]);

    try {
      ws.current.send(JSON.stringify(messagePayload));
      setInputMessage('');
      setShowEmojiPicker(false);
      setIsTyping(false);
      console.log('Message sent:', messagePayload);
    } catch (error) {
      console.error('Error sending message:', error);
      setError('发送消息失败，请稍后再试');
    }
  }, [inputMessage, isConnected, replayId]);

  const handleInput = (e) => {
    setInputMessage(e.target.value);
  };

  const onEmojiClick = (emojiData, event) => {
    setInputMessage(prevInput => prevInput + emojiData.emoji);
  };

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (typingAreaRef.current && !typingAreaRef.current.contains(event.target)) {
        setIsTyping(false);
        setShowEmojiPicker(false);
      }
    };
    if (isTyping) {
      document.addEventListener('mousedown', handleClickOutside);
    } else {
      document.removeEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isTyping]);

  const handleUndo = () => {
    alert('暂未开放');
  };

  return (
    <div className="play-container">
      {/* 全页面遮罩层 */}
      {showOverlay && (
        <div className="full-page-overlay">
          <button className="play-overlay-btn" onClick={handlePlayClick}>
            <span className="play-icon">▶</span>
          </button>
          {/* 可以在遮罩层上显示视频封面图 */}
          {replay && (
            <div className="overlay-poster" style={{
              backgroundImage: `url(http://localhost:8080/replay/${replay.cover_path})`
            }}></div>
          )}
        </div>
      )}

      {loading ? (
        <div className="loading">加载中...</div>
      ) : error ? (
        <div className="error-message">
          {error}
          <button onClick={handleBackClick} className="back-button" style={{ marginTop: '20px' }}>
            返回首页
          </button>
        </div>
      ) : replay ? (
        <>
          {/* 只有在遮罩层消失后才显示视频播放器 */}
          {!showOverlay && (
            <div
              className="video-player-wrapper"
              onTouchStart={handleTouchStart}
              onTouchMove={handleTouchMove}
              onTouchEnd={handleTouchEnd}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseUp}
            >
              <video
                ref={videoRef}
                className="video-player"
                controls
                autoPlay
                src={`http://localhost:8080/replay/${replay.storage_path}?auth=${token}`}
                poster={`http://localhost:8080/replay/${replay.cover_path}`}
              />

              {/* 拖拽指示器 */}
              {showSeekIndicator && (
                <div className="seek-indicator">
                  <div className={`seek-direction ${dragDirection === 'right' ? 'forward' : 'backward'}`}>
                    {dragDirection === 'right' ? '快进' : '快退'}
                  </div>
                  <div className="seek-time">{formatTime(seekTime)} / {formatTime(videoRef.current?.duration || 0)}</div>
                </div>
              )}

              {/* 播放/暂停状态指示器 */}
              {!isDragging && (
                <div className={`play-pause-indicator ${isPlaying ? 'fade-out' : ''}`}>
                  <span className="play-pause-icon">{isPlaying ? '❚❚' : '▶'}</span>
                </div>
              )}
            </div>
          )}
          <div className="video-controls">
            <button className="back-button" onClick={handleBackClick}>
              返回
            </button>
          </div>
          <div className='stream-panle'>
            <div className="video-discussion-area" ref={messagesEndRef}>
              <ol className="message-box">
                <li>
                  <span className='message-sender'>系统提示</span>
                  <span className='message'>直播内容及互动评论严禁传播违法或不良信息，如有违反，小鹅通将采取封禁措施。严禁未成年人直播或打赏。请谨慎判断，注意财产安全，以防人身或财产损失。</span>
                </li>
                <li>
                  <span className='message-sender'>通知</span>
                  <span className='message'>欢迎进入直播间:<br></br>
                    1、请自行调节手机音量至合适的状态。<br></br>
                    2、直播界面显示讲师发布的内容，听众发言可以在讨论区或以弹幕形式查看。<br></br>
                    3、直播结束后，您可以随时回看全部内容。
                  </span>
                </li>
                {messages.filter((message) => {
                  return message.content != ''
                })
                  .map((message, index) => (
                    <li key={index}>
                      <span className='message-sender'>{message.sender}</span>
                      <span className='message'>{message.content}</span>
                    </li>
                  ))}
              </ol>
            </div>
            {!isTyping && (
              <div className='video-footer'>
                <div className="message-input-area">
                  <input
                    type="text"
                    className="message-input"
                    placeholder="说点什么..."
                    value={inputMessage}
                    onClick={() => {
                      setIsTyping(true);
                    }}
                    onChange={handleInput}
                    readOnly={isTyping}
                  />
                </div>
                {/* 将 emoji 按钮移到这里，使其在 isTyping 为 false 时也可见，但点击后展开输入区域 */}
                <div className='emoji-btn' onClick={() => {
                  setIsTyping(true);
                  setShowEmojiPicker(!showEmojiPicker);
                }}>😀</div>
                <div className='more-btn' onClick={handleUndo}>📞</div>
                <div className='more-btn' onClick={handleUndo}>🎁</div>
                <div className='more-btn' onClick={handleUndo}>👍</div>
              </div>
            )}
            {isTyping && (
              <div className='message-input-typing' ref={typingAreaRef}>
                <input
                  type="text"
                  className="input-typing"
                  placeholder="说点什么~"
                  value={inputMessage}
                  onChange={handleInput}
                  autoFocus
                />
                <div className='emoji-btn-typing' onClick={() => setShowEmojiPicker(!showEmojiPicker)}>😀</div>
                {showEmojiPicker && (
                  <div className="emoji-picker-container">
                    <Picker onEmojiClick={onEmojiClick} />
                  </div>
                )}
                <button className="send-btn" onClick={handleSendMessage}>发送</button>
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="error-message">
          未找到视频
          <button onClick={handleBackClick} className="back-button" style={{ marginTop: '20px' }}>
            返回首页
          </button>
        </div>
      )}
    </div>
  );
}

export default PlayVideo;