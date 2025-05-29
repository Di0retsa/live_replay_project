import { useState, useEffect, useRef } from 'react';
import '../styles/Login.css';

function Login({ onLoginSuccess }) {
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [agreed, setAgreed] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [isButtonActive, setIsButtonActive] = useState(false);
  const [loginType, setLoginType] = useState('code');

  const [captchaVisible, setCaptchaVisible] = useState(false);
  const [captchaData, setCaptchaData] = useState(null);
  const [sliderPosition, setSliderPosition] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const [captchaVerified, setCaptchaVerified] = useState(false);
  const sliderRef = useRef(null);
  const containerRef = useRef(null);
  const currentPositionRef = useRef(0);

  useEffect(() => {
    if (loginType === 'code') {
      setIsButtonActive(phone !== '' && code !== '');
    } else {
      setIsButtonActive(phone !== '' && password !== '');
    }
  }, [phone, code, password, loginType]);

  const getCaptcha = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/login/captcha', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.msg || `HTTP错误 ${response.status}`);
      }

      const data = await response.json();

      if (data.code === 200) {
        setCaptchaData(data);
        setCaptchaVisible(true);
        setSliderPosition(0);
        setCaptchaVerified(false);

        if (sliderRef.current) {
          sliderRef.current.style.backgroundColor = '#4a90e2';
        }

        return true;
      } else {
        throw new Error(data.msg || '获取验证码失败');
      }
    } catch (error) {
      alert(error.message);
      return false;
    }
  };

  const verifyCaptcha = async (captchaID, x) => {
    try {
      const response = await fetch('http://localhost:8080/user/login/captcha', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ captcha_id: captchaID, x: Math.round(3000 * x / window.innerWidth) }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.msg || `HTTP错误 ${response.status}`);
      }

      const data = await response.json();

      if (data.code === 200) {
        if (sliderRef.current) {
          sliderRef.current.style.backgroundColor = '#52c41a';
          sliderRef.current.style.transition = 'background-color 0.3s';
        }

        setCaptchaVerified(true);

        setTimeout(() => {
          setCaptchaVisible(false);
        }, 1000);

        return true;
      } else {
        setSliderPosition(0);
        throw new Error(data.msg || '验证失败，请重试');
      }
    } catch (error) {
      alert(error.message);
      if (error.message.includes('过期'))
        getCaptcha();
      return false;
    }
  };

  const handleDragStart = (e) => {
    if (e.type === 'mousedown') {
      e.preventDefault();
    }
    setIsDragging(true);
  };

  const handleDrag = (e) => {
    if (!isDragging) return;

    const containerWidth = containerRef.current.clientWidth;
    const sliderWidth = sliderRef.current.clientWidth;
    const maxPosition = containerWidth - sliderWidth;

    let newPosition;
    if (e.type === 'touchmove') {
      const touch = e.touches[0];
      const containerRect = containerRef.current.getBoundingClientRect();
      newPosition = touch.clientX - containerRect.left - sliderWidth / 2;
    } else {
      newPosition = e.clientX - containerRef.current.getBoundingClientRect().left - sliderWidth / 2;
    }

    newPosition = Math.max(0, Math.min(newPosition, maxPosition));
    currentPositionRef.current = newPosition;
    setSliderPosition(newPosition);
  };

  const handleDragEnd = async () => {
    if (!isDragging) return;
    setIsDragging(false);

    if (captchaData) {
      const result = await verifyCaptcha(captchaData.data.captcha_id, Math.round(currentPositionRef.current));
      if (!result) {
        setSliderPosition(0);
        currentPositionRef.current = 0;
      }
    }
  };

  useEffect(() => {
    const handleMouseMove = (e) => handleDrag(e);
    const handleMouseUp = () => handleDragEnd();
    const handleTouchMove = (e) => handleDrag(e);
    const handleTouchEnd = () => handleDragEnd();

    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      document.addEventListener('touchmove', handleTouchMove, { passive: false });
      document.addEventListener('touchend', handleTouchEnd);
    }

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      document.removeEventListener('touchmove', handleTouchMove);
      document.removeEventListener('touchend', handleTouchEnd);
    };
  }, [isDragging]);

  const getCode = async (phone) => {
    try {
      const response = await fetch('http://localhost:8080/user/code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ phone }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.msg || `HTTP错误 ${response.status}`);
      }

      const data = await response.json();

      if (data.code === 200) {
        return { success: true, code: data.data, error: null };
      } else {
        throw new Error(data.msg || '注册失败');
      }
    } catch (error) {
      return { success: false, code: -1, error: error.message };
    }
  }

  const verifyCode = async (phone, code) => {
    try {
      const response = await fetch('http://localhost:8080/user/login/code', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ phone, code }),
      })
      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.msg || `HTTP错误 ${response.status}`);
      }

      const data = await response.json();

      if (data.code === 200) {
        onLoginSuccess({
          phone: phone,
          token: data.data.token,
          username: data.data.username || phone
        });
      } else {
        throw new Error(data.msg || '登录失败');
      }
    } catch (error) {
      alert(error.message);
    }
  }

  const verifyPassword = async (phone, password) => {
    try {
      const response = await fetch('http://localhost:8080/user/login/password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ phone, password }),
      })
      if (!response.ok) {
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.msg || `HTTP错误 ${response.status}`);
      }

      const data = await response.json();

      if (data.code === 200) {
        onLoginSuccess({
          phone: phone,
          token: data.data.token,
          username: data.data.username || phone
        });
      } else {
        throw new Error(data.msg || '登录失败');
      }
    } catch (error) {
      alert(error.message);
    }
  }

  const handlePhoneChange = (e) => {
    setPhone(e.target.value);
  };

  const handleCodeChange = (e) => {
    setCode(e.target.value);
  };

  const handlePasswordChange = (e) => {
    setPassword(e.target.value);
  };

  const handleAgreementChange = () => {
    setAgreed(!agreed);
  };

  const toggleLoginType = () => {
    setLoginType(loginType === 'code' ? 'password' : 'code');
  };

  const getVerificationCode = async () => {
    if (!phone) {
      alert('请输入手机号');
      return;
    }
    if (!/^1\d{10}$/.test(phone)) {
      alert('请输入正确的手机号');
      return;
    }

    if (!captchaVerified) {
      const captchaResult = await getCaptcha();
      if (!captchaResult) {
        return;
      }
      return;
    }

    let timer = 60;
    setCountdown(timer);

    const interval = setInterval(() => {
      timer--;
      setCountdown(timer);

      if (timer <= 0) {
        clearInterval(interval);
      }
    }, 1000);

    const result = await getCode(phone);

    if (!result.success) {
      alert(result.error);
    } else {
      alert('模拟短信收到验证码：' + result.code);
    }
  };

  const handleOtherLogin = () => {
    alert('暂未开放');
  };

  const handleLogin = () => {
    if (!phone) {
      alert('请输入手机号');
      return;
    }

    if (!/^1\d{10}$/.test(phone)) {
      alert('请输入正确的手机号');
      return;
    }

    if (loginType === 'code') {
      if (!code) {
        alert('请输入验证码');
        return;
      }
      verifyCode(phone, code);
    } else {
      if (!password) {
        alert('请输入密码');
        return;
      }
      verifyPassword(phone, password);
    }

    if (!agreed) {
      alert('请阅读并同意服务协议和隐私政策');
      return;
    }
  };

  return (
    <div className="login-container">
      <div className="login-header">
        <h2>{loginType === 'code' ? '手机号验证码登录' : '手机号密码登录'}</h2>
        <p className="login-subheader">未注册的手机号验证后将自动注册</p>
      </div>

      <div className="login-form">
        <div className="input-group">
          <input
            type="tel"
            placeholder="请输入手机号(无需区号)"
            value={phone}
            onChange={handlePhoneChange}
          />
        </div>

        {loginType === 'code' ? (
          <div className="input-group verification-code">
            <input
              type="text"
              placeholder="请输入验证码"
              value={code}
              onChange={handleCodeChange}
            />
            <button
              onClick={getVerificationCode}
              disabled={countdown > 0}
            >
              {countdown > 0 ? `${countdown}秒后重新获取` : captchaVerified ? '获取验证码' : '点击验证'}
            </button>
          </div>
        ) : (
          <div className="input-group">
            <input
              type="password"
              placeholder="请输入密码"
              value={password}
              onChange={handlePasswordChange}
            />
          </div>
        )}

        {/* 滑块验证码组件 */}
        {captchaVisible && captchaData && (
          <div className="captcha-container">
            <div className="captcha-title">请完成滑块验证</div>
            <div className="captcha-wrapper">
              <div className="captcha-bg" style={{ backgroundImage: `url(${captchaData.data.background_image})` }}>
                <div
                  className="captcha-puzzle"
                  style={{
                    backgroundImage: `url(${captchaData.data.puzzle_image})`,
                    top: `${captchaData.data.puzzle_y / 30}vw`,
                    left: `${sliderPosition}px`
                  }}
                ></div>
              </div>
              <div className="slider-container" ref={containerRef}>
                <div
                  className="slider-handle"
                  ref={sliderRef}
                  style={{ left: `${sliderPosition}px` }}
                  onMouseDown={handleDragStart}
                  onTouchStart={handleDragStart}
                >
                </div>
                <div className="slider-track"></div>
                <div className="slider-text">向右滑动完成验证</div>
              </div>
            </div>
          </div>
        )}

        <div className="agreement">
          <label>
            <input
              type="checkbox"
              checked={agreed}
              onChange={handleAgreementChange}
            />
            <span className="checkmark"></span>
            <span className="agreement-text">
              我已阅读并同意 <a href="#">服务协议</a>、<a href="#">隐私政策</a> 和 <a href="#">商家隐私声明</a>
            </span>
          </label>
        </div>

        <button
          className={`login-btn ${isButtonActive ? 'active' : ''}`}
          onClick={handleLogin}
        >
          登录
        </button>

        <div className="other-login">
          <a href="#" onClick={(e) => { e.preventDefault(); toggleLoginType(); }}>
            {loginType === 'code' ? '密码登录' : '验证码登录'}
          </a>
        </div>
      </div>

      <div className="login-footer">
        <p>其他登录方式</p>
        <p className="email-login">邮箱登录仅支持非中国大陆地区使用</p>
        <div className="login-methods">
          <button className="other-method" onClick={handleOtherLogin}>
            <div className="method">
              <div className="icon email-icon">📧</div>
            </div>
          </button>
          <button className="other-method" onClick={handleOtherLogin}>
            <div className="method">
              <div className="icon wechat-icon">💬</div>
            </div>
          </button>
          <button className="other-method" onClick={handleOtherLogin}>
            <div className="method">
              <div className="icon qq-icon">🔔</div>
            </div>
          </button>
        </div>
      </div>
    </div>
  );
}

export default Login;