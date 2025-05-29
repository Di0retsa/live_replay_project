import React, { useState, useEffect, useRef } from 'react';
import axios from 'axios';

const SliderCaptcha = ({ onSuccess, onFail, onRefresh }) => {
  const [captchaId, setCaptchaId] = useState('');
  const [backgroundImage, setBackgroundImage] = useState('');
  const [puzzleImage, setPuzzleImage] = useState('');
  const [puzzleY, setPuzzleY] = useState(0); // 拼图块的垂直位置

  const [isLoading, setIsLoading] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [sliderLeft, setSliderLeft] = useState(0); // 滑块的当前 left 值
  const [startX, setStartX] = useState(0); // 拖动开始时的鼠标 X 坐标
  const [startLeft, setStartLeft] = useState(0); // 拖动开始时滑块的 left 值

  const puzzleRef = useRef(null); // 拼图块的引用
  const containerRef = useRef(null); // 容器的引用，用于获取宽度

  const fetchCaptcha = async () => {
    setIsLoading(true);
    try {
      // 替换为你的后端 API 地址
      const response = await axios.get('http://localhost:8080/user/login/captcha');
      const data = response.data;
      setCaptchaId(data.captcha_id);
      setBackgroundImage(data.background_image);
      setPuzzleImage(data.puzzle_image);
      setPuzzleY(data.puzzle_y);
      setSliderLeft(0);
      setIsLoading(false);
      if (onRefresh) onRefresh();
    } catch (error) {
      console.error('获取验证码失败:', error);
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchCaptcha();
  }, []);

  const handleMouseDown = (e) => {
    if (isLoading) return;
    setIsDragging(true);
    setStartX(e.clientX || e.touches[0].clientX);
    setStartLeft(sliderLeft);
  };

  const handleMouseMove = (e) => {
    if (!isDragging || isLoading) return;
    const currentX = e.clientX || e.touches[0].clientX;
    let offset = currentX - startX;
    let newLeft = startLeft + offset;

    const containerWidth = containerRef.current ? containerRef.current.offsetWidth : 0;
    const puzzleWidth = puzzleRef.current ? puzzleRef.current.offsetWidth : 60;

    // 限制边界
    if (newLeft < 0) newLeft = 0;
    if (newLeft > containerWidth - puzzleWidth) {
      newLeft = containerWidth - puzzleWidth;
    }
    setSliderLeft(newLeft);
  };

  const handleMouseUp = async () => {
    if (!isDragging || isLoading) return;
    setIsDragging(false);

    // 发送验证请求
    try {
      setIsLoading(true);
      // 替换为你的后端 API 地址
      const response = await axios.post('http://localhost:8080/user/login/captcha', {
        captcha_id: captchaId,
        x: Math.round(sliderLeft),
      });
      if (response.status === 200) {
        if (onSuccess) onSuccess();
      }
    } catch (error) {
      console.error('验证失败:', error);
      if (onFail) onFail();
      fetchCaptcha();
    } finally {
      setIsLoading(false);
    }
  };

  // 确保在 document 上监听 mousemove 和 mouseup，以防止鼠标移出元素范围后事件丢失
  useEffect(() => {
    const handleGlobalMouseMove = (e) => {
      if (isDragging) {
        handleMouseMove(e);
      }
    };
    const handleGlobalMouseUp = () => {
      if (isDragging) {
        handleMouseUp();
      }
    };

    if (typeof window !== 'undefined') {
      window.addEventListener('mousemove', handleGlobalMouseMove);
      window.addEventListener('mouseup', handleGlobalMouseUp);
      window.addEventListener('touchmove', handleGlobalMouseMove, { passive: false });
      window.addEventListener('touchend', handleGlobalMouseUp);
    }


    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('mousemove', handleGlobalMouseMove);
        window.removeEventListener('mouseup', handleGlobalMouseUp);
        window.removeEventListener('touchmove', handleGlobalMouseMove);
        window.removeEventListener('touchend', handleGlobalMouseUp);
      }
    };
  }, [isDragging, startX, startLeft]);


  const puzzlePieceStyle = {
    position: 'absolute',
    top: `${puzzleY}px`, // 与后端返回的位置一致
    left: `${sliderLeft}px`,
    width: '60px', // 与后端裁剪的宽度一致
    height: '60px', // 与后端裁剪的高度一致
    cursor: isDragging ? 'grabbing' : 'grab',
    backgroundImage: `url(${puzzleImage})`,
    backgroundSize: 'cover', // 或者 'contain', '100% 100%'
    zIndex: 2,
    userSelect: 'none', // 防止拖动时选中文本
  };

  const containerStyle = {
    position: 'relative',
    width: '300px',
    height: '150px',
    margin: '20px auto',
    border: '1px solid #ccc',
    backgroundImage: `url(${backgroundImage})`,
    backgroundSize: 'cover',
    userSelect: 'none',
    overflow: 'hidden',
  };

  return (
    <div>
      <div
        ref={containerRef}
        style={containerStyle}
        onTouchStart={handleMouseDown}
      >
        {backgroundImage && puzzleImage && (
          <>
            {/* 背景图的挖孔部分已经在后端处理 */}
            <div
              ref={puzzleRef}
              style={puzzlePieceStyle}
              onMouseDown={handleMouseDown}
            />
          </>
        )}
        {isLoading && <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)' }}>加载中...</div>}
      </div>
      {!isLoading && <p style={{ textAlign: 'center', userSelect: 'none' }}>拖动上方滑块完成验证</p>}
      <button onClick={fetchCaptcha} disabled={isLoading}>
        {isLoading ? '加载中...' : '刷新验证码'}
      </button>
    </div>
  );
};

export default SliderCaptcha;