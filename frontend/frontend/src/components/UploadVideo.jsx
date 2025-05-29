import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import '../styles/UploadVideo.css';
import api from '../utils/api';

function UploadVideo() {
  const navigate = useNavigate();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [videoFile, setVideoFile] = useState(null);
  const [coverFile, setCoverFile] = useState(null);
  const [videoPreview, setVideoPreview] = useState('');
  const [coverPreview, setCoverPreview] = useState('');
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState('');

  const handleVideoChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setVideoFile(file);
      const videoUrl = URL.createObjectURL(file);
      setVideoPreview(videoUrl);
    }
  };

  const handleCoverChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      setCoverFile(file);
      const imageUrl = URL.createObjectURL(file);
      setCoverPreview(imageUrl);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!title.trim()) {
      setError('请输入视频标题');
      return;
    }

    if (!description.trim()) {
      setError('请输入视频描述');
      return;
    }

    if (!videoFile) {
      setError('请选择要上传的视频文件');
      return;
    }

    if (!coverFile) {
      setError('请选择视频封面图片');
      return;
    }

    setUploading(true);
    setUploadProgress(0);
    setError('');

    try {
      const token = localStorage.getItem('token');
      if (!token) {
        throw new Error('未登录，请先登录');
      }

      const formData = new FormData();
      formData.append('title', title);
      formData.append('description', description);
      formData.append('uploadVideo', videoFile);
      formData.append('uploadCover', coverFile);

      const progressInterval = setInterval(() => {
        setUploadProgress(prev => {
          if (prev >= 95) {
            clearInterval(progressInterval);
            return 95;
          }
          const newProgress = prev + Math.random() * 5;
          return Math.min(newProgress, 95);
        });
      }, 500);

      const response = await api.post('/replay/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      });

      clearInterval(progressInterval);
      setUploadProgress(99);

      const data = response.data;

      if (data.code === 200) {
        setUploadProgress(100);

        setTimeout(() => {
          alert('视频上传成功！');
          navigate('/');
        }, 500);
      } else {
        throw new Error(data.msg || '上传失败');
      }
    } catch (error) {
      setError(error.message);
      setUploading(false);
    }
  };

  const handleCancel = () => {
    navigate('/');
  };

  const handleDescriptionChange = (e) => {
    const input = e.target.value;
    if (input.length <= 200) {
      setDescription(input);
    }
  };

  return (
    <div className="upload-container">
      <header className="upload-header">
        <h2>上传视频</h2>
        <button className="cancel-btn" onClick={handleCancel}>取消</button>
      </header>

      {error && <div className="error-message">{error}</div>}

      {uploading && (
        <div className="upload-overlay">
          <div className="upload-modal">
            <div className="spinner"></div>
            <h3>正在上传中，请稍等</h3>
            <div className="progress-bar">
              <div
                className="progress-fill"
                style={{ width: `${Math.min(uploadProgress, 100)}%` }}
              ></div>
            </div>
            <p>{Math.round(uploadProgress)}%</p>
          </div>
        </div>
      )}

      <form className="upload-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <label>视频标题</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="请输入视频标题"
            maxLength="50"
          />
        </div>

        <div className="form-group">
          <label>视频描述</label>
          <textarea
            value={description}
            onChange={handleDescriptionChange}
            placeholder="请输入视频描述"
            maxLength="200"
            rows="4"
          />
          <div className="char-counter">
            <span className={description.length >= 180 ? 'near-limit' : ''}>
              {description.length}/200
            </span>
          </div>
        </div>

        <div className="upload-section">
          <div className="upload-video">
            <label>上传视频</label>
            <input
              type="file"
              accept="video/*"
              onChange={handleVideoChange}
              className="file-input"
              id="video-upload"
            />
            <label htmlFor="video-upload" className="file-label">
              {videoFile ? videoFile.name : '选择视频文件'}
            </label>
            {videoPreview && (
              <div className="preview video-preview">
                <video src={videoPreview} controls></video>
              </div>
            )}
          </div>

          <div className="upload-cover">
            <label>上传封面</label>
            <input
              type="file"
              accept="image/*"
              onChange={handleCoverChange}
              className="file-input"
              id="cover-upload"
            />
            <label htmlFor="cover-upload" className="file-label">
              {coverFile ? coverFile.name : '选择封面图片'}
            </label>
            {coverPreview && (
              <div className="preview cover-preview">
                <img src={coverPreview} alt="视频封面预览" />
              </div>
            )}
          </div>
        </div>

        <button
          type="submit"
          className="submit-btn"
          disabled={uploading}
        >
          {uploading ? '上传中...' : '提交'}
        </button>
      </form>
    </div>
  );
}

export default UploadVideo;